// module-template-go — Botmother tashqi modul namunasi (Go).
//
// Node turlari (SDK imkoniyatlari namunasi):
//   - mymodule.Echo       — action: matnni echo qiladi
//   - mymodule.AuthHeader — action: credential'dan HTTP header (Result.Alerts namuna)
//   - mymodule.OnKeyword  — trigger: kalit so'z (Global toggle)
//   - mymodule.Route      — action: NOMLI dinamik chiqishlar (found/missing)
//   - mymodule.PickSheet  — action: KASKADLI dynamic_select (credential→hujjat→varaq)
//
// Modul o'z credential turini ham e'lon qiladi (mymodule.apikey).
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
		Color:       "integration-sky",
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
	// AI TOOL — modul AI agent uchun DINAMIK tool beradi.
	//   Builder AI agent node'ga "mymodule.search" tool'ini accessory qilib
	//   ulaydi; LLM uni chaqirsa, engine shu modulga JSON-RPC yuboradi va
	//   Invoke qaytargan matn LLM'ga natija bo'ladi.
	// ------------------------------------------------------------------
	m.AddTool(botmodule.Tool{
		Name:        "mymodule.search",
		Description: "Katalogdan mahsulot qidiradi va topilganlarni qaytaradi.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Qidiruv so'rovi (mahsulot nomi yoki kalit so'z).",
				},
			},
			"required": []any{"query"},
		},
		Invoke: func(c *botmodule.ToolCtx) (string, error) {
			query := c.String("query")
			if query == "" {
				return "", fmt.Errorf("query bo'sh bo'lishi mumkin emas")
			}
			// Bu yerda real qidiruv (DB/API) bo'lardi — namuna uchun statik javob.
			return fmt.Sprintf("'%s' bo'yicha 3 ta mahsulot topildi: Kitob A, Kitob B, Kitob C", query), nil
		},
	})

	// ------------------------------------------------------------------
	// Modul o'z CREDENTIAL TURINI e'lon qiladi — foydalanuvchi shu turdan
	// credential yaratadi. Input soni/turini modul o'zi belgilaydi.
	// ------------------------------------------------------------------
	m.AddCredentialType(botmodule.CredentialType{
		Key:   "mymodule.apikey",
		Label: "MyModule API",
		Icon:  "key",
		Color: "#10A37F",
		Mode:  "header", // engine bu credential'ni qanday qo'llashi
		Fields: []botmodule.CredentialField{
			{Name: "token", Label: "API Token", Type: "text", Required: true, Secret: true},
			{Name: "base_url", Label: "Base URL", Type: "text", Placeholder: "https://api.example.com"},
			{Name: "model", Label: "Model", Type: "select", Options: []botmodule.SelectOption{
				{Value: "small", Label: "Small"},
				{Value: "large", Label: "Large"},
			}},
		},
	})

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
		Color:       "action-emerald",
		Content: []botmodule.Field{
			{
				Type:           "credential",
				Key:            "api_credential",
				Label:          "Credential",
				CredentialType: "mymodule.apikey", // modul e'lon qilgan tur bilan filtr
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
					ContextUpdates: map[string]any{"auth_header": "", "cred_type": ""},
					ExitOutput:     "error",
					Alerts: []botmodule.Alert{{
						Level:   botmodule.AlertError,
						Message: "credential tanlanmagan",
					}},
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
		Color:       "trigger-blue",
		Trigger:     true,
		TriggerMode: "event-match",
		Global:      true, // har qanday holatda ishlasin (boshqa flow ichida bo'lsa ham)
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

	// ------------------------------------------------------------------
	// 4. Action node — Route (NOMLI DINAMIK CHIQISHLAR namunasi)
	//    Qiymatga qarab "found" yoki "missing" chiqishiga yo'naltiradi.
	// ------------------------------------------------------------------
	m.AddNode(botmodule.Node{
		Type:        "mymodule.Route",
		Title:       "Yo'naltirish (Route)",
		Description: "Qiymat bo'lsa 'found', bo'lmasa 'missing' chiqishiga ketadi",
		Category:    "integrations",
		Icon:        "split",
		Color:       "integration-indigo",
		Content: []botmodule.Field{
			{Type: "text", Key: "value", Label: "Qiymat", Placeholder: "{{state.user_id}}"},
		},
		Defaults: map[string]any{"value": ""},
		Outputs: []botmodule.Output{
			{Name: "found", Label: "Topildi", Variant: "success"},
			{Name: "missing", Label: "Topilmadi", Variant: "danger"},
		},
		Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
			if strings.TrimSpace(c.String("value")) != "" {
				return botmodule.Result{
					ExitOutput:     "found",
					ContextUpdates: map[string]any{"route_value": c.String("value")},
				}
			}
			return botmodule.Result{ExitOutput: "missing"}
		},
	})

	// ------------------------------------------------------------------
	// 5. Action node — PickSheet (KASKADLI dynamic_select namunasi)
	//    Credential → hujjat → varaq (Google Sheets uslubida). Varaq ro'yxati
	//    tanlangan hujjatga (dependsOn) bog'liq.
	// ------------------------------------------------------------------
	m.AddNode(botmodule.Node{
		Type:        "mymodule.PickSheet",
		Title:       "Jadval tanlash",
		Description: "Credential → hujjat → varaq (kaskadli dinamik tanlov)",
		Category:    "integrations",
		Icon:        "table",
		Color:       "integration-green",
		Content: []botmodule.Field{
			{Type: "credential", Key: "api_credential", Label: "Credential", CredentialType: "mymodule.apikey"},
			{Type: "dynamic_select", Key: "doc_id", Label: "Hujjat", Resource: "docs", CredentialKey: "api_credential"},
			{Type: "dynamic_select", Key: "sheet_id", Label: "Varaq", Resource: "sheets", CredentialKey: "api_credential", DependsOn: []string{"doc_id"}},
		},
		ProducesState: []string{"picked_doc", "picked_sheet"},
		Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
			return botmodule.Result{ContextUpdates: map[string]any{
				"picked_doc":   c.String("doc_id"),
				"picked_sheet": c.String("sheet_id"),
			}}
		},
	})

	// dynamic_select options loaderlari — DEMO ma'lumot. Haqiqiy modulda
	// c.Credential() bilan tashqi API'ga so'rov yuborib ro'yxat olinadi.
	m.AddOptionsLoader("docs", func(c *botmodule.OptionsCtx) []botmodule.SelectOption {
		return []botmodule.SelectOption{
			{Value: "doc_1", Label: "Hisobot 2026"},
			{Value: "doc_2", Label: "Mijozlar"},
		}
	})
	m.AddOptionsLoader("sheets", func(c *botmodule.OptionsCtx) []botmodule.SelectOption {
		// dependsOn: tanlangan hujjatga qarab varaqlar.
		switch c.String("doc_id") {
		case "doc_1":
			return []botmodule.SelectOption{{Value: "s_yan", Label: "Yanvar"}, {Value: "s_fev", Label: "Fevral"}}
		case "doc_2":
			return []botmodule.SelectOption{{Value: "s_active", Label: "Faol"}, {Value: "s_archive", Label: "Arxiv"}}
		}
		return nil
	})

	// ------------------------------------------------------------------
	// 6. Action node — PickCity (CREDENTIAL'SIZ kaskadli dynamic_select)
	//    Davlat → shahar. Credential YO'Q (field'larda credentialKey berilmagan)
	//    — optionlar modul tomonidan auth'siz hisoblanadi.
	// ------------------------------------------------------------------
	m.AddNode(botmodule.Node{
		Type:        "mymodule.PickCity",
		Title:       "Shahar tanlash",
		Description: "Credential'siz kaskadli tanlov (davlat → shahar)",
		Category:    "integrations",
		Icon:        "globe",
		Color:       "integration-sky",
		Content: []botmodule.Field{
			// credentialKey YO'Q → credential talab qilinmaydi
			{Type: "dynamic_select", Key: "country", Label: "Davlat", Resource: "countries"},
			{Type: "dynamic_select", Key: "city", Label: "Shahar", Resource: "cities", DependsOn: []string{"country"}},
		},
		ProducesState: []string{"picked_country", "picked_city"},
		Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
			return botmodule.Result{ContextUpdates: map[string]any{
				"picked_country": c.String("country"),
				"picked_city":    c.String("city"),
			}}
		},
	})

	// Credential'siz options loaderlari — auth kerak emas, faqat dependsOn.
	m.AddOptionsLoader("countries", func(c *botmodule.OptionsCtx) []botmodule.SelectOption {
		return []botmodule.SelectOption{
			{Value: "uz", Label: "O'zbekiston"},
			{Value: "kz", Label: "Qozog'iston"},
		}
	})
	m.AddOptionsLoader("cities", func(c *botmodule.OptionsCtx) []botmodule.SelectOption {
		switch c.String("country") {
		case "uz":
			return []botmodule.SelectOption{{Value: "tashkent", Label: "Toshkent"}, {Value: "samarkand", Label: "Samarqand"}}
		case "kz":
			return []botmodule.SelectOption{{Value: "almaty", Label: "Almati"}, {Value: "astana", Label: "Ostona"}}
		}
		return nil
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
