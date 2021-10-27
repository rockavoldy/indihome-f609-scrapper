package main

import (
	"errors"
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const admin_url = "http://192.168.1.1"
const username = "admin"
const password = "Telkomdso123"

func main() {
	launcher := launcher.New().Headless(false)
	defer launcher.Cleanup()

	url := launcher.MustLaunch()
	browser := rod.New().ControlURL(url).Trace(true).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(admin_url)

	err := login(page)
	if err != nil {
		log.Fatalln(err)
	}

	iframe_template := page.MustElement("iframe[src='template.gch']").MustFrame()
	IPAddress, err := getIPAddress(iframe_template)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(IPAddress)
}

func login(page *rod.Page) error {
	page.MustElement("input[id='Frm_Username']").MustInput(username)
	page.MustElement("input[id='Frm_Password']").MustInput(password)
	page.MustElement("input[id='LoginId']").MustClick()

	if exist, errmsg, err := page.Has("font[id='errmsg']"); exist || err != nil {
		return errors.New(errmsg.String())
	}

	return nil
}

func getIPAddress(page *rod.Page) (string, error) {
	// logout := page.MustElementR("a", "Logout")
	// // defer logout.MustClick()

	network_info := page.MustElement("table.menu_table tr.h2_content")
	network_info.MustClick()

	IPAddress := page.MustElement("input[id='TextPPPIPAddress0']").MustAttribute("value")
	return *IPAddress, nil
}
