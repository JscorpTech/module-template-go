# module-template-go

Botmother tashqi moduli — Go template.

Bu repo yangi modul yaratish uchun boshlang'ich nuqta. `botmodule-go` SDK'ni
ishlatadi va uchta node turi ko'rsatadi: Echo (action), AuthHeader (credential),
OnKeyword (trigger).

---

## Tez boshlash

```bash
git clone https://github.com/yourorg/module-template-go
cd module-template-go
```

**O'zgartirish kerak bo'lgan joylar:**

1. `module.yaml` — `module.id` (global unique slug), `name`, `author`, `github`.
2. `main.go` — `moduleID` konstanta va barcha `"mymodule.*"` qiymatlarni
   o'z modulId'ingizga almashtiring.
3. Node'larni o'zgartiring yoki yangisini qo'shing.

```bash
# Lokal test
go run .

# Boshqa terminaldagi test
curl http://localhost:8100/health
curl -X POST http://localhost:8100/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"describe","id":1}'
```

---

## Modul yaratish qadamlari

### 1. ID va nom o'zgartirish

`module.yaml`:
```yaml
module:
  id: mymodule        # o'zgartiring: global unique slug
  name: My Module     # ko'rinadigan nom
```

`main.go`:
```go
const moduleID = "mymodule"  // module.yaml bilan mos bo'lsin
```

### 2. Node qo'shish

```go
m.AddNode(botmodule.Node{
    Type:     "mymodule.MyNode",      // moduleID.NodeName format
    Title:    "Mening node'im",
    Category: "integrations",
    Icon:     "box",
    Color:    "blue",
    Content: []botmodule.Field{
        {Type: "text", Key: "param1", Label: "Parametr"},
    },
    Defaults: map[string]any{"param1": ""},
    Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
        val := c.String("param1")
        return botmodule.Result{
            ContextUpdates: map[string]any{"result": val},
        }
    },
})
```

### 3. Trigger qo'shish

```go
m.AddNode(botmodule.Node{
    Type:        "mymodule.OnEvent",
    Title:       "Hodisa kelganda",
    Category:    "triggers",
    Trigger:     true,
    TriggerMode: "event-match",
    Content: []botmodule.Field{
        {Type: "text", Key: "pattern", Label: "Pattern"},
    },
    Match: func(c *botmodule.TriggerCtx) botmodule.MatchResult {
        text := c.MessageText()
        pattern := c.String("pattern")
        return botmodule.MatchResult{
            Matched: strings.Contains(text, pattern),
        }
    },
})
```

### 4. Credential ishlatish

```go
// Manifest:
Field{Type: "credential", Key: "api_cred", Label: "API Credential"}

// Handler:
Execute: func(c *botmodule.ExecuteCtx) botmodule.Result {
    cred, ok := c.Credential("api_cred")
    if !ok {
        return botmodule.Result{ContextUpdates: map[string]any{}}
    }
    apiKey := cred.Data["api_key"]
    // ... HTTP so'rov yuboring
}
```

---

## SDK qanday ulanadi

`botmodule-go` SDK alohida public repo: `github.com/BotSpace/botmodule-go`.
Template uni oddiy Go dependency sifatida `go.mod`'da `require` qiladi —
build paytida GitHub'dan yuklanadi (`go mod download`). Vendoring yo'q.

SDK versiyasini yangilash:
```bash
go get github.com/BotSpace/botmodule-go@latest
go mod tidy
git add go.mod go.sum && git commit -m "bump SDK" && git push
```

> `go.sum` commit qilinadi — build checksum'ni o'sha fayldan tekshiradi
> (Dockerfile'da `GOSUMDB=off`, shu sabab sum.golang.org kerak emas).

## Lokal test (Docker)

```bash
docker build -t mymodule .
docker run -p 8100:8100 mymodule
curl http://localhost:8100/health
```

---

## Qisqa ma'lumotnoma (cheat sheet)

To'liq tafsilot — [`SDK.md`](./SDK.md). Tez-tez kerak bo'ladiganlar shu yerda:

### Field `type` (SDK orqali to'liq ishlaydigani)

| `type` | UI | Qo'shimcha xossa | `data[key]` qiymati |
|---|---|---|---|
| `text` | input | `placeholder`, `helpText` | string |
| `number` | raqamli input | `placeholder` | number |
| `textarea` | ko'p qatorli | `placeholder` | string |
| `select` | dropdown | **`options[]`** | `option.value` |
| `checkbox` | katakcha | — | bool |
| `switch` | toggle | `helpText` | bool |
| `credential` | credential tanlovchi | **`credentialType`** | engine `credentials[key]` ga uzatadi |

Boshqa type'lar (array-editor, nested-rows, code-editor, ...) qo'shimcha props talab
qiladi — SDK `Field` ularni hozir bermaydi. Shu 7 tasi bilan ishlang.

### Ikon (`Icon`) — eng kerakli nomlar

`zap` `globe` `http` `database` `key` `brain-circuit` `table` `folder` `arrow-right`
`settings-2` `message` `clock` `credit-card` `code`. Noma'lum nom → `sparkles`.
To'liq ro'yxat SDK.md §8.

### Rang (`Color`) — token kalitlari

Action: `integration-sky` `integration-green` `action-violet` `action-emerald`.
Trigger: `trigger-blue`. AI: `ai-violet`. **Noma'lum token → kulrang** (`"blue"` ishlamaydi!).
To'liq ro'yxat SDK.md §9.

### `visibleWhen` (shartli ko'rsatish)

```go
{Type: "switch", Key: "save", Optional: true},
{Type: "text", Key: "state_key", Optional: true,
 VisibleWhen: &botmodule.VisibleWhen{Key: "save", Equals: true}},
```

### Handler kontekstidan o'qish

```go
c.String("key")          // data[key] → string
c.Int("key")             // data[key] → int64
c.Credential("key")      // (*Credential, bool)
c.Data["flag"].(bool)    // switch/checkbox
c.MessageText()          // trigger: update.message.text
c.CallbackData()         // trigger: callback_query.data
```

### Fayl bilan ishlash (ExecuteCtx)

```go
uuid, _ := c.UploadFile("a.pdf", bytes)                 // doimiy saqlash → UUID (state'ga qo'ying)
uuid, _ = c.UploadFileWithTTL("temp.pdf", bytes, 3600)  // 3600s saqlanadi, keyin avto-o'chadi (0/<0 = doimiy)
data, _ := c.GetFile(uuid)                              // o'qish → []byte
_ = c.DeleteFile(uuid)                                  // o'chirish
```
Engine fayl API'ni avtomatik beradi (project'ga scoped). TTL bilan saqlangan fayllar
muddati o'tgach platforma tomonidan avtomatik tozalanadi. Batafsil: SDK.md §20.5.

### Boshqa yangi imkoniyatlar (SDK.md §20)

- **`Result{Alerts: []Alert{...}}`** — info/warning/error xabarni debug UI'da ko'rsatadi (§20.1)
- **`Node.Outputs`** — nomli dinamik chiqishlar; `Result{ExitOutput:"found"}` (§20.2)
- **`m.AddCredentialType(...)`** — modul o'z credential turini beradi (§20.3)
- **`dynamic_select` + `m.AddOptionsLoader(...)`** — kaskadli tanlov, doc→sheet (§20.4)
- **`Node.Global:true`** — global trigger toggle (§20.6)

---

## SDK to'liq hujjati

Bu repodagi **[`SDK.md`](./SDK.md)** — `botmodule-go` ning to'liq ma'lumotnomasi
(har bir tur, field type, ikon va rang token'lari, JSON-RPC kontrakti, `module.yaml`,
test). Packagega borish shart emas — hammasi shu repoda.

Manba: `github.com/BotSpace/botmodule-go`.
