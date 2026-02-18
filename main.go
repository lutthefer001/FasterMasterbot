package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/telebot.v3"
)

// --- KONFIGURATSIYA ---
const (
	// MUHIM: @BotFather dan olingan yangi tokenni shu yerga qo'ying.
	BotToken   = "8424913938:AAH_gr6L8c1UMFyCXubApgdsB0ZbGtKXbrw"
	WeatherKey = "fba0dd4cca49587c78358e9fd0bcbd1e"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	pref := telebot.Settings{
		Token:  BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("🚨 Bot ishga tushmadi! Tokenni tekshiring: %v", err)
	}

	// --- ASOSIY MENYU (8 TA TUGMA) ---
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
	
	btnMoney   := menu.Text("💵 Valyuta Kurslari")
	btnWeather := menu.Text("🌦 Ob-havo")
	btnQR      := menu.Text("🔳 QR Kod")
	btnPass    := menu.Text("🔑 Parol")
	btnWiki    := menu.Text("📖 Wikipedia")
	btnID      := menu.Text("🆔 Mening ID")
	btnInsta   := menu.Text("📸 Instagram")
	btnYT      := menu.Text("📺 YouTube")

	menu.Reply(
		menu.Row(btnMoney, btnWeather),
		menu.Row(btnQR, btnPass),
		menu.Row(btnWiki, btnID),
		menu.Row(btnInsta, btnYT),
	)

	// --- FUNKSIYALAR ---

	// 1. Start
	b.Handle("/start", func(c telebot.Context) error {
		return c.Send("Salom!Men hamma narsanni qila oladigan botman. Menyudan foydalaning:", menu)
	})

	// 2. Valyuta Kursi
	b.Handle(&btnMoney, func(c telebot.Context) error {
		c.Send("🔄 Kurslar yangilanmoqda...")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("https://cbu.uz/uz/arkhiv-kursov-valyut/json/")
		if err != nil {
			return c.Send("❌ Markaziy Bank bilan ulanishda xatolik.")
		}
		defer resp.Body.Close()

		var rates []struct {
			CcyNm_UZ string `json:"CcyNm_UZ"`
			Code     string `json:"Ccy"`
			Rate     string `json:"Rate"`
		}
		json.NewDecoder(resp.Body).Decode(&rates)

		msg := "💰 **MB Kurslari:**\n\n"
		important := map[string]bool{"USD": true, "EUR": true, "RUB": true}
		for _, r := range rates {
			if important[r.Code] {
				msg += fmt.Sprintf("🏳️ %s: `%s` UZS\n", r.CcyNm_UZ, r.Rate)
			}
		}
		return c.Send(msg, telebot.ModeMarkdown)
	})

	// 3. Ob-havo
	b.Handle(&btnWeather, func(c telebot.Context) error {
		return c.Send("Ob-havo uchun `/w shahar` deb yozing (Masalan: `/w Toshkent`)")
	})
	b.Handle("/w", func(c telebot.Context) error {
		if len(c.Args()) == 0 { return c.Send("Shahar nomini kiriting!") }
		apiURL := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s&lang=uz", url.QueryEscape(strings.Join(c.Args(), " ")), WeatherKey)
		resp, _ := http.Get(apiURL)
		if resp.StatusCode != 200 { return c.Send("❌ Topilmadi.") }
		var d struct {
			Main struct { Temp float64 `json:"temp"` } `json:"main"`
			Name string `json:"name"`
		}
		json.NewDecoder(resp.Body).Decode(&d)
		return c.Send(fmt.Sprintf("🌤 %s: %.1f°C", d.Name, d.Main.Temp))
	})

	// 4. QR Kod
	b.Handle(&btnQR, func(c telebot.Context) error {
		return c.Send("QR kod yaratish uchun `/qr matn` deb yozing.")
	})
	b.Handle("/qr", func(c telebot.Context) error {
		if len(c.Args()) == 0 { return c.Send("Matn kiriting!") }
		qr := "https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=" + url.QueryEscape(strings.Join(c.Args(), " "))
		return c.Send(&telebot.Photo{File: telebot.FromURL(qr)})
	})

	// 5. Parol yaratish
	b.Handle(&btnPass, func(c telebot.Context) error {
		chars := "abcdefghijklmnopqrstuvwxyz0123456789!@"
		p := make([]byte, 10)
		for i := range p { p[i] = chars[rand.Intn(len(chars))] }
		return c.Send(fmt.Sprintf("🔑 Yangi parol: `%s`", string(p)), telebot.ModeMarkdown)
	})

	// 6. Wikipedia
	b.Handle(&btnWiki, func(c telebot.Context) error {
		return c.Send("Wikipedia'dan qidirish uchun `/wiki mavzu` deb yozing.")
	})
	b.Handle("/wiki", func(c telebot.Context) error {
		if len(c.Args()) == 0 { return c.Send("Mavzu kiriting!") }
		return c.Send("📖 Wikipedia: https://uz.wikipedia.org/wiki/" + url.QueryEscape(strings.Join(c.Args(), " ")))
	})

	// 7. Mening ID-m
	b.Handle(&btnID, func(c telebot.Context) error {
		return c.Send(fmt.Sprintf("🆔 Sizning ID: `%d` \n👤 Ism: %s", c.Sender().ID, c.Sender().FirstName), telebot.ModeMarkdown)
	})

	// 8. Instagram & YouTube yuklovchi (Linklarni ushlash)
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		txt := c.Text()
		if strings.Contains(txt, "instagram.com") {
			c.Send("⏳ Instagram video yuklanmoqda...")
			dl := strings.Replace(txt, "instagram.com", "ddinstagram.com", 1)
			return c.Send(&telebot.Video{File: telebot.FromURL(dl)})
		}
		if strings.Contains(txt, "youtube.com") || strings.Contains(txt, "youtu.be") {
			return c.Send("📺 YouTube yuklash uchun havola: https://ssyoutube.com/uz/\n\nSiz yuborgan link: " + txt)
		}
		return nil
	})

	b.Handle(&btnInsta, func(c telebot.Context) error { return c.Send("Menga Instagram video linkini yuboring.") })
	b.Handle(&btnYT, func(c telebot.Context) error { return c.Send("Menga YouTube video linkini yuboring.") })

	log.Println("✅ 8 ta funksiyali bot ishga tushdi!")
	b.Start()

}

