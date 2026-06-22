// module-template-go — Botmother tashqi modul namunasi (Go).
//
// Uchta node turi ko'rsatiladi:
//   - mymodule.Echo       — action: matnni echo qiladi
//   - mymodule.AuthHeader — action: credential'dan HTTP header quradi
//   - mymodule.OnKeyword  — trigger: xabar matnida kalit so'z bo'lsa fire bo'ladi
//
// Yangi modul yozish uchun:
//  1. module.yaml: module.id va name o'zgartiring (slugni ham o'zgartiring).
//  2. Shu fayldagi "mymodule" ni o'z modulId'ingizga almashtiring.
//  3. Node'larni qo'shing yoki o'chiring.
//  4. Serve porti 8100 (PORT env orqali ham o'zgartiriladi).
package main

import (
	"fmt"
	"strings"

	botmodule "github.com/BotSpace/botmodule-go"
)

const moduleID = "mymodule"

func main() {
	m := botmodule.New(moduleID, "My Module")
	m.Version = "0.1.0"
	m.Docs = docs

	// ------------------------------------------------------------------
	// 1. Action node — Echo
	//    Kiritilgan matnni echo_output o'zgaruvchisiga yozadi.
	// ------------------------------------------------------------------
	m.AddNode(botmodule.Node{
		Type:        "mymodule.Echo",
		Title:       "Echo",
		Description: "Kiritilgan matnni o'zgaruvchiga yozadi",
		Category:    "integrations",
		Icon:        "sparkles",
		Color:       "blue",
		Content: []botmodule.Field{
			{
				Type:        "text",
				Key:         "input",
				Label:       "Matn",
				Placeholder: "{{message.text}} yoki literal",
				HelpText:    "Natija echo_output o'zgaruvchisiga yoziladi",
			},
		},
		Defaults:      map[string]any{"input": "{{message.text}}"},
		ProducesState: []string{"echo_output"},
		Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
			return botmodule.Result{
				ContextUpdates: map[string]any{
					"echo_output": c.String("input"),
				},
			}
		},
	})

	// ------------------------------------------------------------------
	// 2. Action node — AuthHeader (credential ishlatish namunasi)
	//    Tanlangan credential'dan Authorization header quradi.
	//    Engine credential_id ni o'zi resolve qilib decrypted sirni uzatadi.
	// ------------------------------------------------------------------
	m.AddNode(botmodule.Node{
		Type:        "mymodule.AuthHeader",
		Title:       "Auth header",
		Description: "Credential'dan HTTP auth header quradi",
		Category:    "integrations",
		Icon:        "credit-card",
		Color:       "emerald",
		Content: []botmodule.Field{
			{
				Type:           "credential",
				Key:            "api_credential",
				Label:          "Credential",
				CredentialType: "", // bo'sh = har qanday tip qabul qilinadi
			},
			{
				Type:        "text",
				Key:         "note",
				Label:       "Izoh",
				Placeholder: "ixtiyoriy",
				Optional:    true,
			},
		},
		Defaults:      map[string]any{"api_credential": "", "note": ""},
		ProducesState: []string{"auth_header", "cred_type"},
		Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
			cred, ok := c.Credential("api_credential")
			if !ok {
				return botmodule.Result{
					ContextUpdates: map[string]any{
						"auth_header": "",
						"cred_type":   "",
					},
				}
			}

			var header string
			switch cred.Mode {
			case "bearer":
				header = "Bearer " + cred.Data["token"] + cred.Data["api_key"]
			case "basic":
				user := cred.Data["username"]
				pass := cred.Data["password"]
				header = fmt.Sprintf("Basic <base64(%s:%s)>", user, pass)
			case "header":
				header = cred.Data["value"]
			default:
				if v := cred.Data["api_key"]; v != "" {
					header = v
				} else {
					header = cred.Data["token"]
				}
			}

			// Sirni to'liq oqizmaslik: maskalash (demo amaliyoti).
			masked := header
			if len(masked) > 16 {
				masked = masked[:16] + "..."
			}

			return botmodule.Result{
				ContextUpdates: map[string]any{
					"auth_header": masked,
					"cred_type":   cred.TypeKey,
				},
			}
		},
	})

	// ------------------------------------------------------------------
	// 3. Trigger node — OnKeyword (event-match)
	//    Xabar matnida kalit so'z bo'lsa flow'ni ishga tushiradi.
	// ------------------------------------------------------------------
	m.AddNode(botmodule.Node{
		Type:        "mymodule.OnKeyword",
		Title:       "Kalit so'z kelganda",
		Description: "Xabarda berilgan kalit so'z bo'lsa ishga tushadi",
		Category:    "triggers",
		Icon:        "zap",
		Color:       "amber",
		Trigger:     true,
		TriggerMode: "event-match",
		Content: []botmodule.Field{
			{
				Type:        "text",
				Key:         "keyword",
				Label:       "Kalit so'z",
				Placeholder: "salom",
				HelpText:    "Katta-kichik harf farq qilmaydi",
			},
		},
		Defaults: map[string]any{"keyword": ""},
		Match: func(c *botmodule.TriggerCtx) botmodule.MatchResult {
			text := c.MessageText()
			kw := strings.TrimSpace(c.String("keyword"))
			if kw == "" {
				return botmodule.MatchResult{Matched: false}
			}
			matched := strings.Contains(strings.ToLower(text), strings.ToLower(kw))
			updates := map[string]any{}
			if matched {
				updates["matched_keyword"] = kw
			}
			return botmodule.MatchResult{Matched: matched, ContextUpdates: updates}
		},
	})

	m.Serve(":8100")
}

// ------------------------------------------------------------------
// Modul hujjati (markdown) — admin panel va docs() metodi uchun.
// ------------------------------------------------------------------
const docs = `# My Module

Botmother tashqi modul namunasi (Go). JSON-RPC 2.0 kontrakti.

## Node turlari

### ` + "`mymodule.Echo`" + ` (action)

Kiritilgan matnni ` + "`echo_output`" + ` o'zgaruvchisiga yozadi.

- **input**: matn (` + "`{{message.text}}`" + ` yoki literal)
- **Chiqish**: ` + "`echo_output`" + `

### ` + "`mymodule.AuthHeader`" + ` (action, credential)

Tanlangan credential'dan HTTP Authorization header quradi.

- **api_credential**: credential picker (` + "`credential`" + ` field)
- **Chiqish**: ` + "`auth_header`" + ` (maskalangan), ` + "`cred_type`" + `

### ` + "`mymodule.OnKeyword`" + ` (trigger, event-match)

Telegram xabar matnida kalit so'z bo'lsa flow'ni ishga tushiradi.

- **keyword**: kalit so'z (katta-kichik harf farq qilmaydi)
- **Chiqish**: ` + "`matched_keyword`" + ` o'zgaruvchisi

## Misol flow

` + "```" + `
OnKeyword (keyword: salom)
  → Echo (input: {{message.text}})
  → Matn yuborish ({{echo_output}})
` + "```" + `

Bot "salom" so'zli xabar kelganda ishlaydi va matnni qaytaradi.
`
