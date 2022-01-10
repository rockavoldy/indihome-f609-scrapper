package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/joho/godotenv"
)

type Page struct {
	page      *rod.Page
	admin_url string
	username  string
	password  string
}

func NewPage(page *rod.Page, admin_url, username, password string) *Page {
	return &Page{
		page:      page,
		admin_url: admin_url,
		username:  username,
		password:  password,
	}
}

func (p *Page) Login() error {
	time.Sleep(200 * time.Millisecond)
	p.page.MustElement("input[id='Frm_Username']").MustInput(p.username)
	p.page.MustElement("input[id='Frm_Password']").MustInput(p.password)
	p.page.MustElement("input[id='LoginId']").MustClick()

	if exist, errmsg, err := p.page.Has("font[id='errmsg']"); exist || err != nil {
		return errors.New(errmsg.String())
	}

	return nil
}

func (p *Page) WANInfoPage() *rod.Page {
	network_info := p.page.MustElement("table.menu_table tr.h2_content")
	network_info.MustClick()

	return network_info.Page()
}

func (p *Page) GetIPAddress(ipAddress chan string) (string, error) {
	IPAddress := p.page.MustElement("input[id='TextPPPIPAddress0']").MustAttribute("value")

	ipAddress <- *IPAddress
	return *IPAddress, nil
}

func (p *Page) GetConnStatus(connStatus chan string) (string, error) {
	ConnStatus := p.page.MustElement("input[id='TextPPPConStatus0']").MustAttribute("value")

	connStatus <- *ConnStatus
	return *ConnStatus, nil
}

func (p *Page) Logout(signal chan bool) {
	if err := p.page.MustElement("div[class='title_log'] > a").Click(proto.InputMouseButtonLeft); err != nil {
		log.Println(err)
	}

	signal <- true
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("There is no .env found in the directory, please add!")
	}

	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	admin_url := os.Getenv("ADMIN_URL")

	launcher := launcher.New().Headless(true)
	defer launcher.Cleanup()

	url := launcher.MustLaunch()
	browser := rod.New().ControlURL(url).Trace(true).MustConnect()
	defer browser.MustClose()

	page := NewPage(browser.MustPage(admin_url), admin_url, username, password)

	err = page.Login()
	if err != nil {
		log.Fatalln(err)
	}

	bodyPage := NewPage(page.page.MustElement("iframe[src='template.gch']").MustFrame(), admin_url, username, password)

	connInfoPage := NewPage(bodyPage.WANInfoPage(), admin_url, username, password)

	IPAddress := make(chan string, 1)
	ConnStatus := make(chan string, 1)
	go func() {
		_, err = connInfoPage.GetIPAddress(IPAddress)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = connInfoPage.GetConnStatus(ConnStatus)
		if err != nil {
			log.Fatalln(err)
		}

	}()

	log.Println("IP Address: ", <-IPAddress)
	log.Println("Status: ", <-ConnStatus)

	signalSuccessLogout := make(chan bool, 1)

	go bodyPage.Logout(signalSuccessLogout)

	log.Println("Logout? ", <-signalSuccessLogout)
}
