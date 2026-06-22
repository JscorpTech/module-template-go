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

## SDK `vendor/` ichida — push qilishdan oldin yangilang

SDK kodi repo ichidagi `vendor/` papkada keladi. Shu sabab platforma (yoki
Kaniko) reponi klonlab, **network'siz** to'g'ridan-to'g'ri build qiladi —
`go get` yoki tashqi proxy kerak emas.

Node'larni o'zgartirgach `vendor/` ni yangilang va push qiling:

```bash
go mod tidy
go mod vendor      # SDK'ni vendor/ ga ko'chiradi
git add -A && git commit -m "update" && git push
```

> `go.mod` dagi `replace ... => ../botmodule-go` faqat SDK'ni parallel ishlab
> chiqayotganlar uchun. `vendor/` mavjud bo'lgani uchun build uni o'qimaydi —
> o'chirishingiz shart emas.

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

---

## SDK to'liq hujjati

Bu repodagi **[`SDK.md`](./SDK.md)** — `botmodule-go` ning to'liq ma'lumotnomasi
(har bir tur, field type, ikon va rang token'lari, JSON-RPC kontrakti, `module.yaml`,
test). Packagega borish shart emas — hammasi shu repoda.

Manba: `github.com/BotSpace/botmodule-go`.
