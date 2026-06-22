module github.com/BotSpace/module-template-go

go 1.22

require github.com/BotSpace/botmodule-go v0.1.0

// Lokal dev uchun: SDK'ni parallel ishlab chiqayotganingizda ishlatiladi.
// Prodga push qilishdan avval bu qatorni olib tashlang.
replace github.com/BotSpace/botmodule-go => ../botmodule-go
