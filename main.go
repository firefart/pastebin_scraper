package main

import (
  "encoding/json"
  "crypto/tls"
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"
  "sync"
  "time"
  "os/signal"
  "archive/zip"
  "bytes"
  "math/rand"
  "path"
  "regexp"

  "gopkg.in/gomail.v2"
)

const (
  apiEndpoint = "https://scrape.pastebin.com/api_scraping.php"
  userAgent   = "Pastebin Scraper (https://firefart.at)"
)

var (
  debug      = flag.Bool("debug", false, "Print debug output")
  configFile = flag.String("config", "", "Config File to use")

  config configuration

  lastCheck time.Time
  terminate = false
  alredyChecked = make(map[string]time.Time)

  client = &http.Client{
    Timeout: 10 * time.Second,
  }

  chanError  = make(chan error)
  chanOutput = make(chan paste)
  chanSignal = make(chan os.Signal)

  wgOutput sync.WaitGroup
  wgError  sync.WaitGroup

  r = rand.New(rand.NewSource(time.Now().UnixNano()))

  keywordsRegex = make(map[string]*regexp.Regexp)
)

type configuration struct {
  Mailserver  string    `json:"mailserver"`
  Mailport    int       `json:"mailport"`
  Mailfrom    string    `json:"mailfrom"`
  Mailonerror bool      `json:"mailonerror"`
  Mailtoerror string    `json:"mailtoerror"`
  Mailto      string    `json:"mailto"`
  Mailsubject string    `json:"mailsubject"`
  Keywords    []string  `json:"keywords"`
}

type paste struct {
  FullURL           string  `json:"full_url"`
  ScrapeURL         string  `json:"scrape_url"`
  Date              string  `json:"date"`
  Key               string  `json:"key"`
  Size              string  `json:"size"`
  Expire            string  `json:"expire"`
  Title             string  `json:"title"`
  Syntax            string  `json:"syntax"`
  User              string  `json:"user"`
  Content           string
  MatchedKeywords   map[string]string
}

func randomString(strlen int) string {
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
  result := make([]byte, strlen)
  for i := range result {
    result[i] = chars[r.Intn(len(chars))]
  }
  return string(result)
}

func pasteToPrettyString(p paste) string {
  keywords := strings.Join(getKeysFromMap(p.MatchedKeywords), ",")
  var buffer bytes.Buffer
  buffer.WriteString(fmt.Sprintf("Pastebin Alert for Keywords %s\n\n", keywords))
  if p.Title != "" {
    buffer.WriteString(fmt.Sprintf("Title: %s\n", p.Title))
  }
  if p.FullURL != "" {
    buffer.WriteString(fmt.Sprintf("URL: %s\n", p.FullURL))
  }
  if p.User != "" {
    buffer.WriteString(fmt.Sprintf("User: %s\n", p.User))
  }
  if p.Date != "" {
    buffer.WriteString(fmt.Sprintf("Date: %s\n", p.Date))
  }
  if p.Size != "" {
    buffer.WriteString(fmt.Sprintf("Size: %s\n", p.Size))
  }
  if p.Expire != "" {
    buffer.WriteString(fmt.Sprintf("Expire: %s\n", p.Expire))
  }
  if p.Syntax != "" {
    buffer.WriteString(fmt.Sprintf("Syntax: %s\n", p.Syntax))
  }

  for k, v := range p.MatchedKeywords {
    buffer.WriteString(fmt.Sprintf("\nFirst match of Keyword: %s\n%s", k, v))
  }

  return buffer.String()
}

func createZip(filename string, content string) (zipContent []byte, err error) {
  buf := new(bytes.Buffer)
  w := zip.NewWriter(buf)
  f, err := w.Create(filename)
  if err != nil {
    return
  }
  if _, err = f.Write([]byte(content)); err != nil {
    return
  }
  if err = w.Close(); err != nil {
    return
  }
  return buf.Bytes(), nil
}

func debugOutput(s string) {
  if *debug {
    log.Printf("[DEBUG] %s", s)
  }
}

func httpRequest(url string) (*http.Response, error) {
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return nil, err
  }
  req.Header.Set("User-Agent", userAgent)

  resp, err := client.Do(req)
  return resp, err
}

func httpRespBodyToString(resp *http.Response) (res string, err error) {
  if resp == nil {
    return "", fmt.Errorf("Response is nil")
  }

  // catch errors when closing and return them
  defer func() {
    rerr := resp.Body.Close()
    if rerr != nil {
      err = rerr
    }
  }()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }

  res = string(body)
  return res, nil
}

func fetchPasteList() ([]paste, error) {
  var list []paste
  debugOutput("Fetching paste list")
  url := fmt.Sprintf("%s?limit=100", apiEndpoint)
  resp, err := httpRequest(url)
  if err != nil {
    return list, err
  }

  body, err := httpRespBodyToString(resp)
  if err != nil {
    return list, err
  }
  if strings.Contains(body, "DOES NOT HAVE ACCESS") {
    panic("You do not have access to the scrape API from this IP address!")
  }

  lastCheck = time.Now()
  jsonErr := json.Unmarshal([]byte(body), &list)
  if jsonErr != nil {
    return list, fmt.Errorf("Error on parsing JSON: %v. JSON: %s", jsonErr, body)
  }
  return list, nil
}

func sendEmail(m *gomail.Message) (error) {
  debugOutput("Sending Mail")
  d := gomail.Dialer{Host: config.Mailserver, Port: config.Mailport}
  d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gas
  err := d.DialAndSend(m)
  return err
}

func getKeysFromMap(in map[string]string) []string {
    keys := make([]string, 0, len(in))
    for k := range in {
      keys = append(keys, k)
    }
    return keys
}

func sendPasteMessage(p paste) (err error) {
  m := gomail.NewMessage()
  m.SetHeader("From", config.Mailfrom)
  m.SetHeader("To", config.Mailto)
  keywords := strings.Join(getKeysFromMap(p.MatchedKeywords), ",")
  m.SetHeader("Subject", fmt.Sprintf("Pastebin Alert for %s", keywords))

  filename := fmt.Sprintf("%s.zip", randomString(10))
  fullPath := path.Join(os.TempDir(), filename)
  zipFile, err := createZip("content.txt", p.Content)
  if err != nil {
    return err
  }

  f, err := os.Create(fullPath)
  if err != nil {
    return err
  }
  defer func() {
    rerr := os.Remove(fullPath)
    if rerr != nil {
      err = rerr
    }
  }()

  _, err = f.Write(zipFile)
  if err != nil {
    return err
  }

  defer func() {
    rerr := f.Close()
    if rerr != nil {
      err = rerr
    }
  }()

  m.Attach(fullPath)

  body := pasteToPrettyString(p)
  m.SetBody("text/plain", body)
  err = sendEmail(m)
  return err
}

func sendErrorMessage(errorMessage error) error {
  m := gomail.NewMessage()
  m.SetHeader("From", config.Mailfrom)
  m.SetHeader("To", config.Mailtoerror)
  m.SetHeader("Subject", "ERROR in pastebin_scraper")
  m.SetBody("text/plain", fmt.Sprintf("%v", errorMessage))

  err := sendEmail(m)
  return err
}

func checkKeywords(body string) (bool, map[string]string) {
  found := make(map[string]string)
  status := false
  for k, v := range keywordsRegex {
    if v != nil {
      if s := v.FindStringSubmatch(body); s != nil {
        found[k] = strings.TrimSpace(s[1])
        status = true
      }
    }
  }
  return status, found
}

func fetch(p paste) (error) {
  debugOutput(fmt.Sprintf("Checking Paste %s", p.Key))
  alredyChecked[p.Key] = time.Now()
  resp, err := httpRequest(p.ScrapeURL)
  if err != nil {
    return err
  }

  if resp.StatusCode == http.StatusOK || resp.ContentLength > 0 {
    b, err := httpRespBodyToString(resp)
    if err != nil {
      return err
    }
    found, key := checkKeywords(b)
    if found {
      p.Content = b
      p.MatchedKeywords = key
      chanOutput <- p
    }
  } else {
    b, err := httpRespBodyToString(resp)
    return fmt.Errorf("Output: %s, Error: %v", b, err)
  }
  return nil
}

// nolint: gocyclo
func main() {
  flag.Parse()

  if *configFile == "" {
    log.Fatalln("Please provide a valid config file")
  }

  file, err := os.Open(*configFile)
  if err != nil {
    log.Fatalf("Error opening config file: %v", err)
  }

  defer func() {
    rerr := file.Close()
    if rerr != nil {
      log.Fatalf("Error closing config file: %v", rerr)
    }
  }()

  decoder := json.NewDecoder(file)
  config = configuration{}
  err = decoder.Decode(&config)
  if err != nil {
    log.Fatalf("Error parsing config file: %v", err)
  }

  log.Println("Starting Pastebin Scraper")

  // Clean exit when pressing CTRL+C
  signal.Notify(chanSignal, os.Interrupt)
  go func() {
    for range chanSignal {
      fmt.Println("Detected CTRL+C. Please wait a sec for exit")
      terminate = true
    }
  }()

  wgOutput.Add(1)
  wgError.Add(1)

  go func() {
    defer wgOutput.Done()
    for p := range chanOutput {
      debugOutput(fmt.Sprintf("Found Paste:\n%s", pasteToPrettyString(p)))
      err = sendPasteMessage(p)
      if err != nil {
        chanError <- fmt.Errorf("sendPasteMessage: %v", err)
      }
    }
  }()

  go func() {
    defer wgError.Done()
    for err := range chanError {
      log.Printf("%v", err)
      if config.Mailonerror {
        err2 := sendErrorMessage(err)
        if err2 != nil {
          log.Printf("ERROR on sending error mail: %v", err2)
        }
      }
    }
  }()

  // use a boundary for keyword searching
  for _, k := range config.Keywords {
    r := fmt.Sprintf(`(?im)^(.*\b%s.*)$`, regexp.QuoteMeta(k))
    keywordsRegex[k] = regexp.MustCompile(r)
  }

  for {
    if terminate {
      break
    }

    // Only fetch the main list once a minute
    sleepTime := time.Until(lastCheck.Add(1 * time.Minute))
    if sleepTime > 0 {
      debugOutput(fmt.Sprintf("Sleeping for %s", sleepTime))
      time.Sleep(sleepTime)
    }

    pastes, err := fetchPasteList()
    if err != nil {
      chanError <- fmt.Errorf("fetchPasteList: %v", err)
    }

    for _, p := range pastes {
      if terminate {
        break
      }
      if _, ok := alredyChecked[p.Key]; ok {
        debugOutput(fmt.Sprintf("Skipping key %s as it was already checked", p.Key))
      } else {
        err := fetch(p)
        if err != nil {
          chanError <- fmt.Errorf("fetch: %v", err)
        }
        // do not hammer the API
        time.Sleep(1 * time.Second)
      }
    }
    // clean up old items in alreadyChecked map
    // delete everything older than 10 minutes
    threshold := time.Now().Add(-10 * time.Minute)
    for k, v := range alredyChecked {
      if v.Before(threshold) {
        debugOutput(fmt.Sprintf("Deleting expired entry %s", k))
        delete(alredyChecked, k)
      }
    }
  }

  close(chanOutput)
  wgOutput.Wait()
  close(chanError)
  wgError.Wait()
}
