package main

import (
  "encoding/json"
  "crypto/tls"
  "errors"
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
  apiEndpoint = "https://pastebin.com/api_scraping.php"
  userAgent   = "Pastebin Scraper (https://firefart.at)"
)

var (
  debug  = flag.Bool("debug", false, "Print debug output")
  config = flag.String("config", "", "Config File to use")

  configuration Configuration

  lastCheck time.Time
  terminate = false
  alredyChecked = make(map[string]time.Time)

  client = &http.Client{
    Timeout: time.Duration(10 * time.Second),
  }

  chanError  = make(chan error)
  chanOutput = make(chan Paste)
  chanSignal = make(chan os.Signal)

  wgOutput sync.WaitGroup
  wgError  sync.WaitGroup

  r = rand.New(rand.NewSource(time.Now().UnixNano()))

  keywordsRegex = make(map[string]*regexp.Regexp)
)

type Configuration struct {
  Mailserver  string
  Mailport    int
  Mailfrom    string
  Mailtoerror string
  Mailto      string
  Mailsubject string
  Keywords    []string
  Errorfile   string
}

type Paste struct {
  Full_url          string
  Scrape_url        string
  Date              string
  Key               string
  Size              string
  Expire            string
  Title             string
  Syntax            string
  User              string
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

func pasteToPrettyString(paste Paste) string {
  keywords := strings.Join(getKeysFromMap(paste.MatchedKeywords), ",")
  var buffer bytes.Buffer
  buffer.WriteString(fmt.Sprintf("Pastebin Alert for Keywords %s\n\n", keywords))
  if paste.Title != "" {
    buffer.WriteString(fmt.Sprintf("Title: %s\n", paste.Title))
  }
  if paste.Full_url != "" {
    buffer.WriteString(fmt.Sprintf("URL: %s\n", paste.Full_url))
  }
  if paste.User != "" {
    buffer.WriteString(fmt.Sprintf("User: %s\n", paste.User))
  }
  if paste.Date != "" {
    buffer.WriteString(fmt.Sprintf("Date: %s\n", paste.Date))
  }
  if paste.Size != "" {
    buffer.WriteString(fmt.Sprintf("Size: %s\n", paste.Size))
  }
  if paste.Expire != "" {
    buffer.WriteString(fmt.Sprintf("Expire: %s\n", paste.Expire))
  }
  if paste.Syntax != "" {
    buffer.WriteString(fmt.Sprintf("Syntax: %s\n", paste.Syntax))
  }

  for k, v := range paste.MatchedKeywords {
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

func httpRespBodyToString(resp *http.Response) (string, error) {
  if resp == nil {
    return "", errors.New("Response is nil")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }

  b := string(body)
  return b, nil
}

func fetchPasteList() (list []Paste, err error) {
  debugOutput("Fetching paste list")
  url := fmt.Sprintf("%s?limit=100", apiEndpoint)
  resp, err := httpRequest(url)
  if err != nil {
    return
  }
  jsonErr := json.NewDecoder(resp.Body).Decode(&list)
  lastCheck = time.Now()
  if jsonErr != nil {
    b, err := httpRespBodyToString(resp)
    if err != nil {
      b = err.Error()
    }
    err = errors.New(fmt.Sprintf("Error on parsing JSON: %s. JSON: %s", jsonErr.Error(), b))
  }
  return
}

func sendEmail(m *gomail.Message) (err error) {
  debugOutput("Sending Mail")
  d := gomail.Dialer{Host: configuration.Mailserver, Port: configuration.Mailport}
  d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
  err = d.DialAndSend(m)
  return
}

func getKeysFromMap(in map[string]string) []string {
    keys := make([]string, 0, len(in))
    for k := range in {
      keys = append(keys, k)
    }
    return keys
}

func sendPasteMessage(paste Paste) (err error) {
  m := gomail.NewMessage()
  m.SetHeader("From", configuration.Mailfrom)
  m.SetHeader("To", configuration.Mailto)
  keywords := strings.Join(getKeysFromMap(paste.MatchedKeywords), ",")
  m.SetHeader("Subject", fmt.Sprintf("Pastebin Alert for %s", keywords))

  filename := fmt.Sprintf("%s.zip", randomString(10))
  fullPath := path.Join(os.TempDir(), filename)
  zipFile, err := createZip("content.txt", paste.Content)
  if err != nil {
    return
  }

  f, err := os.Create(fullPath)
  if err != nil {
    return
  }
  defer os.Remove(fullPath)
  f.Write(zipFile)
  f.Close()
  m.Attach(fullPath)

  body := pasteToPrettyString(paste)
  m.SetBody("text/plain", body)
  err = sendEmail(m)
  return
}

func sendErrorMessage(errorString string) error {
  m := gomail.NewMessage()
  m.SetHeader("From", configuration.Mailfrom)
  m.SetHeader("To", configuration.Mailtoerror)
  m.SetHeader("Subject", "ERROR in pastebin_scraper")
  m.SetBody("text/plain", errorString)

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

func fetch(paste Paste) (error) {
  debugOutput(fmt.Sprintf("Checking Paste %s", paste.Key))
  alredyChecked[paste.Key] = time.Now()
  resp, err := httpRequest(paste.Scrape_url)
  if err != nil {
    return err
  }

  if resp.StatusCode == 200 || resp.ContentLength > 0 {
    b, err := httpRespBodyToString(resp)
    if err != nil {
      return err
    }
    found, key := checkKeywords(b)
    if found {
      paste.Content = b
      paste.MatchedKeywords = key
      chanOutput <- paste
    }
  } else {
    b, err := httpRespBodyToString(resp)
    if err != nil {
      b = err.Error()
    }
    return errors.New(fmt.Sprintf("Output: %s, Error: %s", b, err.Error()))
  }
  return nil
}

func main() {
  flag.Parse()

  if *config == "" {
    log.Fatalln("Please provide a valid config file")
  }

  file, err := os.Open(*config)
  if err != nil {
    log.Fatalf("Error opening config file: %s", err.Error())
  }
  defer file.Close()
  decoder := json.NewDecoder(file)
  configuration = Configuration{}
  err = decoder.Decode(&configuration)
  if err != nil {
    log.Fatalf("Error parsing config file: %s", err.Error())
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
    for paste := range chanOutput {
      debugOutput(fmt.Sprintf("Found Paste:\n%s", pasteToPrettyString(paste)))
      err = sendPasteMessage(paste)
      if err != nil {
        chanError <- errors.New(fmt.Sprintf("sendPasteMessage: %s", err.Error()))
      }
    }
  }()

  go func() {
    defer wgError.Done()
    for err := range chanError {
      log.Println(err.Error())
      // ignore errors
      sendErrorMessage(err.Error())
    }
  }()

  // use a boundary for keyword searching
  for _, k := range configuration.Keywords {
    r := fmt.Sprintf(`(?im)^(.*\b%s.*)$`, regexp.QuoteMeta(k))
    keywordsRegex[k] = regexp.MustCompile(r)
  }

  for {
    if terminate {
      break
    }

    // Only fetch the main list once a minute
    sleepTime := lastCheck.Add(1 * time.Minute).Sub(time.Now())
    if sleepTime > 0 {
      debugOutput(fmt.Sprintf("Sleeping for %s", sleepTime))
      time.Sleep(sleepTime)
    }

    pastes, err := fetchPasteList()
    if err != nil {
      chanError <- errors.New(fmt.Sprintf("fetchPasteList: %s", err.Error()))
    }

    for _, paste := range pastes {
      if terminate {
        break
      }
      if _, ok := alredyChecked[paste.Key]; ok {
        debugOutput(fmt.Sprintf("Skipping key %s as it was already checked", paste.Key))
      } else {
        err := fetch(paste)
        if err != nil {
          chanError <- errors.New(fmt.Sprintf("fetch: %s", err.Error()))
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
