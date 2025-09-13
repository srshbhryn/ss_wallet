package enums

type PayoutStatus string

const NEW = PayoutStatus("new")
const SENT = PayoutStatus("sent")
const SUCCESS = PayoutStatus("success")
const FAILED = PayoutStatus("failed")

type BankType string

const DUMMY = BankType("dummy")
const SAMANAN = BankType("saman")
const MELLAT = BankType("mellat")
