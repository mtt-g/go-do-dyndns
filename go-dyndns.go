package main

import (
  "context"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "time"
  "github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

const (
    pat = "[Access Token Here]"
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
  tokenSource := &TokenSource{
		AccessToken: pat,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

  publicIp := getPublicIp("https://api.ipify.org")
  searchDomain := os.Args[1]
  searchSubdomain := os.Args[2]

  ctx := context.TODO()

  dnsRecordId := getDNSRecordId(ctx, client, searchDomain, searchSubdomain)

  if dnsRecordId > 0 {
    updateDNSRecord(ctx, client, searchDomain, dnsRecordId, publicIp)
    log.Printf("IP address successfully updated to %s", publicIp)
  }

}

func getDNSRecordId(ctx context.Context, client *godo.Client, searchDomain string, searchSubdomain string) int {

  subDomainId := 0

  domains := getDNSRecords(ctx, client, searchDomain)

  for _, subdomain := range domains {
    if subdomain.Name == searchSubdomain {
      log.Printf("IP Address assigned to DNS record: %s", subdomain.Data)
      subDomainId = subdomain.ID
    }
  }

  return subDomainId
}

func getDNSRecords(ctx context.Context, client *godo.Client, searchDomain string) []godo.DomainRecord {

  domains, _, err := client.Domains.Records(ctx, searchDomain, nil)

  if err != nil {
    log.Fatalln(err)
  }

  return domains
}

func updateDNSRecord(ctx context.Context, client *godo.Client, domain string, id int, ip string) {

  _, _, err := client.Domains.EditRecord(ctx, domain, id, &godo.DomainRecordEditRequest{Data: ip})

  if err != nil {
    log.Fatalln(err)
  }
}

func getPublicIp(url string) string {

  var netClient = http.Client{
    Timeout: time.Second * 5,
  }

  resp, err := netClient.Get(url)

  if err != nil {
    log.Fatalln(err)
  }

  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalln(err)
  }

  return string(body)
}
