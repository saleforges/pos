package domain

import (
	"fmt"
	"strings"
	"time"
)

// ReceiptWidth is the printable character width of a 58mm thermal
// receipt (most common for UMKM POS).
const ReceiptWidth = 32

// Receipt is the printable struk payload for an order.
type Receipt struct {
	OrderID   int64            `json:"orderId"`
	BranchID  int64            `json:"branchId"`
	Status    string           `json:"status"`
	Total     float64          `json:"total"`
	Paid      float64          `json:"paid"`
	Change    float64          `json:"change"`
	Debt      float64          `json:"debt"`
	Items     []ReceiptItem    `json:"items"`
	Payments  []ReceiptPayment `json:"payments"`
	CreatedAt time.Time        `json:"createdAt"`
	Text      string           `json:"text"`
}

type ReceiptItem struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Total    float64 `json:"total"`
}

type ReceiptPayment struct {
	Method string    `json:"method"`
	Amount float64   `json:"amount"`
	PaidAt time.Time `json:"paidAt"`
}

// BuildReceipt renders the printable struk for an order.
func BuildReceipt(o *Order) *Receipt {
	r := &Receipt{
		OrderID:   o.ID,
		BranchID:  o.BranchID,
		Status:    string(o.Status),
		Total:     o.Total,
		Paid:      o.PaidAmount,
		CreatedAt: o.CreatedAt,
		Items:     make([]ReceiptItem, 0, len(o.Items)),
		Payments:  make([]ReceiptPayment, 0, len(o.Payments)),
	}
	for _, it := range o.Items {
		r.Items = append(r.Items, ReceiptItem{
			Name:     it.ItemName,
			Quantity: it.Quantity,
			Price:    it.UnitPrice,
			Total:    it.LineTotal,
		})
	}
	for _, p := range o.Payments {
		r.Payments = append(r.Payments, ReceiptPayment{
			Method: string(p.Method),
			Amount: p.Amount,
			PaidAt: p.PaidAt,
		})
	}
	if o.PaidAmount > o.Total {
		r.Change = o.PaidAmount - o.Total
	} else if o.PaidAmount < o.Total {
		r.Debt = o.Total - o.PaidAmount
	}
	r.Text = r.render()
	return r
}

func (r *Receipt) render() string {
	var b strings.Builder
	sep := strings.Repeat("-", ReceiptWidth)
	line := func(s string) { b.WriteString(s + "\n") }
	center := func(s string) {
		if len(s) >= ReceiptWidth {
			line(s)
			return
		}
		pad := (ReceiptWidth - len(s)) / 2
		line(strings.Repeat(" ", pad) + s)
	}
	kv := func(k, v string) {
		line(k + strings.Repeat(".", ReceiptWidth-len(k)-len(v)) + v)
	}

	center("SALEFORGES")
	center("Struk Transaksi")
	line(sep)
	kv("Order", fmt.Sprintf("#%d", r.OrderID))
	kv("Cabang", fmt.Sprintf("%d", r.BranchID))
	kv("Waktu", r.CreatedAt.Format("02/01/2006 15:04"))
	line(sep)
	for _, it := range r.Items {
		name := it.Name
		if len(name) > ReceiptWidth-14 {
			name = name[:ReceiptWidth-14]
		}
		line(name)
		line(fmt.Sprintf("  %g x %s", it.Quantity, money(it.Price)))
		line(fmt.Sprintf("%s%s", strings.Repeat(" ", ReceiptWidth-10), money(it.Total)))
	}
	line(sep)
	kv("Total", money(r.Total))
	kv("Dibayar", money(r.Paid))
	if r.Change > 0 {
		kv("Kembali", money(r.Change))
	}
	if r.Debt > 0 {
		kv("Sisa", money(r.Debt))
	}
	line(sep)
	for _, p := range r.Payments {
		kv("Bayar "+p.Method, money(p.Amount))
	}
	line(sep)
	center("Terima kasih!")
	center("Barang yang sudah dibeli")
	center("tidak dapat dikembalikan")
	return b.String()
}

func money(v float64) string {
	return fmt.Sprintf("Rp%.0f", v)
}
