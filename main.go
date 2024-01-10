package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"

	objs "github.com/SakoDroid/telego/objects"
)

const token string = "6652699656:AAG9eBr3T-uM-RpM6CRCzHFsbT5AFOITp8I"

var chatID int = 0

type Item struct {
	title string
	price int
	link  string
}

func (i Item) Equals(other Item) bool {
	return i.link == other.link
}

func (i Item) ToString() string {
	return fmt.Sprintf("Title: %s\nPrice: %d\nLink: %s\n", i.title, i.price, i.link)
}

func scrapeSS() []Item {
	var items []Item

	// web scraping
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.ss.lv/lv/electronics/phones/mobile-phones/apple/", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("authority", "www.ss.lv")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-language", "en-US,en;q=0.9,lv;q=0.8,ru;q=0.7")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("cookie", "LG=lv; sid=05bafd14a9f7dbf4e15d80aa9e74cecc3d1b9124f4562c0275af035a7a4ea43788aab72ec8135c531586c0b77b606221; PHPSESSID=106d8c8359a829ee48c5878535b97b9e; sid_c=1")
	req.Header.Set("if-modified-since", "Wed, 10 Jan 2024 09:27:55 GMT")
	req.Header.Set("referer", "https://www.ss.lv/lv/electronics/phones/mobile-phones/")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Microsoft Edge";v="120"`)
	req.Header.Set("sec-ch-ua-mobile", "?1")
	req.Header.Set("sec-ch-ua-platform", `"Android"`)
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 Edg/120.0.0.0")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	n, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyText := string(n)
	// fmt.Printf("%s\n", bodyText)
	// ==================

	// parse bodyTextand print text from <tr id="tr_ to </tr>
	// var strings []string

	// find subtrsing without regex

	fmt.Println(len(bodyText))

	bodyText = strings.ReplaceAll(bodyText, "\n", "")

	var stringgs []string
	for {
		start := strings.Index(bodyText, "<tr id=\"tr_")
		fmt.Println(start)
		if start == -1 {
			break
		}
		endText := bodyText[start:]
		end := strings.Index(endText, "</tr>")
		end += start + 5
		if start == -1 || end == -1 {
			break
		}
		stringgs = append(stringgs, bodyText[start:end])

		bodyText = bodyText[end:]
	}
	fmt.Println("PARSING")

	var item Item
	i := 0
	for _, s := range stringgs {
		// fmt.Println(s)
		// fmt.Println("=====================================")

		// extract link
		re := regexp.MustCompile("href=.*?>")
		linksObj := re.FindAllStringSubmatch(s, -1)
		if linksObj == nil {
			log.Println("no links")
			continue
		}
		link := linksObj[1][0]
		link = link[6 : len(link)-2]
		link = "https://www.ss.lv" + link
		// fmt.Println(link)

		item.link = link

		// extract title
		re = regexp.MustCompile("\">.*?a>")
		titleObj := re.FindAllStringSubmatch(s, -1)
		if titleObj == nil {
			log.Println("no title")
		}
		title := titleObj[1][0]
		title = title[2 : len(title)-2]
		// fmt.Println(title)

		item.title = title

		// extract price
		re = regexp.MustCompile("c=1>.*?<")
		priceObj := re.FindAllStringSubmatch(s, -1)
		if priceObj == nil {
			log.Println("no price")
		}

		price := priceObj[3][0]
		price = price[4 : len(price)-1]
		// fmt.Println(price)
		price = strings.ReplaceAll(price, " ", "")
		price = strings.ReplaceAll(price, "€", "")

		n, err := strconv.Atoi(price)

		item.price = n

		fmt.Printf("Item %d\nTitle: %s\nPrice: %d\nLink: %s\n\n", i, item.title, item.price, item.link)
		i++
		if err == nil {
			items = append(items, item)
		}
	}

	return items
}

func ssTask(bot *bt.Bot) {
	var oldItems []Item
	var newItems []Item
	for {
		oldItems = newItems
		newItems = scrapeSS()
		if len(oldItems) == 0 {
			oldItems = newItems
		}

		// delete items from newItems that are in oldItems
		for _, oldItem := range oldItems {
			for i, newItem := range newItems {
				if oldItem.Equals(newItem) {
					newItems = append(newItems[:i], newItems[i+1:]...)
				}
			}
		}

		// send newItems to telegram
		if len(newItems) > 0 && chatID != 0 {
			for _, item := range newItems {
				_, err := bot.AdvancedMode().ASendMessage(chatID, item.ToString(), "", 0, false, false, nil, true, false, nil)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		time.Sleep(5 * time.Minute)
	}
}

func main() {
	bot, err := bt.NewBot(cfg.Default(token))

	if err == nil {
		err = bot.Run()
		if err == nil {
			go start(bot)
		}
	}

	ssTask(bot)

	// }

	// fmt.Println(err)
}

func start(bot *bt.Bot) {

	//The general update channel.
	updateChannel := bot.GetUpdateChannel()

	bot.AddHandler("/start", func(u *objs.Update) {

		// //Create the custom keyboard
		// kb := bot.CreateKeyboard(true, false, false, "type ...")
		// //Add buttons to it. First argument is the button's text and the second one is the row number that the button will be added to it.
		// kb.AddButton("button1", 1)
		// kb.AddButton("button2", 1)
		// kb.AddButton("button3", 2)

		chatID = u.Message.Chat.Id

		//Pass the keyboard to the send method
		_, err := bot.AdvancedMode().ASendMessage(u.Message.Chat.Id, "Welcome, to the Facebook Marketplace and SS.lv scraper bot", "", u.Message.MessageId, false, false, nil, true, false, nil)
		if err != nil {
			fmt.Println(err)
		}
	}, "private", "group")

	//Monitores any other update. (Updates that don't contain text message "hi" in a private chat)
	for {
		<-(*updateChannel)

		//Some processing on the update
	}
}
