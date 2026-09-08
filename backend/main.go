package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Contract struct {
	ID            int64   `json:"id"`
	Type          string  `json:"type"`
	Number        string  `json:"number"`
	Name          string  `json:"name"`
	Counterparty  string  `json:"counterparty"`
	Amount        float64 `json:"amount"`
	SettledAmount float64 `json:"settledAmount"`
	PaidAmount    float64 `json:"paidAmount"`
	SignDate      string  `json:"signDate"`
	StartDate     string  `json:"startDate"`
	EndDate       string  `json:"endDate"`
	Status        string  `json:"status"`
	Manager       string  `json:"manager"`
	Notes         string  `json:"notes"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type Settlement struct {
	ID             int64   `json:"id"`
	Type           string  `json:"type"`
	Number         string  `json:"number"`
	ContractNumber string  `json:"contractNumber"`
	ContractName   string  `json:"contractName"`
	Counterparty   string  `json:"counterparty"`
	Period         string  `json:"period"`
	ReportedAmount float64 `json:"reportedAmount"`
	ApprovedAmount float64 `json:"approvedAmount"`
	PaidAmount     float64 `json:"paidAmount"`
	Status         string  `json:"status"`
	Manager        string  `json:"manager"`
	Notes          string  `json:"notes"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type Store struct {
	mu               sync.RWMutex
	Path             string
	NextID           int64        `json:"nextId"`
	Contracts        []Contract   `json:"contracts"`
	NextSettlementID int64        `json:"nextSettlementId"`
	Settlements      []Settlement `json:"settlements"`
}

func loadStore(path string) (*Store, error) {
	s := &Store{Path: path, NextID: 4, Contracts: seedContracts(), NextSettlementID: 3, Settlements: seedSettlements()}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, s.save()
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	s.Path = path
	for _, c := range s.Contracts {
		if c.ID >= s.NextID {
			s.NextID = c.ID + 1
		}
	}
	for _, item := range s.Settlements {
		if item.ID >= s.NextSettlementID {
			s.NextSettlementID = item.ID + 1
		}
	}
	if s.NextSettlementID == 0 {
		s.NextSettlementID = 1
	}
	return s, nil
}
func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, b, 0644)
}
func seedContracts() []Contract {
	now := time.Now().Format(time.RFC3339)
	return []Contract{
		{ID: 1, Type: "owner", Number: "YZ-2026-001", Name: "项目施工总承包合同", Counterparty: "示例建设单位", Amount: 128000000, SettledAmount: 42000000, PaidAmount: 35600000, SignDate: "2026-01-12", StartDate: "2026-02-01", EndDate: "2027-12-31", Status: "履约中", Manager: "张经理", Notes: "示例数据，可直接编辑或删除", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Type: "subcontract", Number: "FB-2026-003", Name: "主体结构劳务分包合同", Counterparty: "示例劳务有限公司", Amount: 26800000, SettledAmount: 9200000, PaidAmount: 7800000, SignDate: "2026-02-18", StartDate: "2026-03-01", EndDate: "2027-03-31", Status: "履约中", Manager: "李工", Notes: "示例数据，可直接编辑或删除", CreatedAt: now, UpdatedAt: now},
		{ID: 3, Type: "subcontract", Number: "FB-2026-008", Name: "机电安装专业分包合同", Counterparty: "示例机电工程有限公司", Amount: 18600000, SettledAmount: 3200000, PaidAmount: 2600000, SignDate: "2026-04-06", StartDate: "2026-04-15", EndDate: "2027-08-31", Status: "履约中", Manager: "王工", CreatedAt: now, UpdatedAt: now},
	}
}
func seedSettlements() []Settlement {
	now := time.Now().Format(time.RFC3339)
	return []Settlement{
		{ID: 1, Type: "owner", Number: "JS-YZ-2026-04", ContractNumber: "YZ-2026-001", ContractName: "项目施工总承包合同", Counterparty: "示例建设单位", Period: "2026-04", ReportedAmount: 12500000, ApprovedAmount: 11800000, PaidAmount: 10500000, Status: "已审核", Manager: "张经理", Notes: "示例数据", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Type: "subcontract", Number: "JS-FB-2026-04", ContractNumber: "FB-2026-003", ContractName: "主体结构劳务分包合同", Counterparty: "示例劳务有限公司", Period: "2026-04", ReportedAmount: 3600000, ApprovedAmount: 3300000, PaidAmount: 2800000, Status: "已审核", Manager: "李工", Notes: "示例数据", CreatedAt: now, UpdatedAt: now},
	}
}

type API struct{ store *Store }

func main() {
	dataPath := env("DATA_FILE", "./data/business.json")
	port := env("PORT", "8080")
	s, err := loadStore(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	api := &API{store: s}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("/api/contracts", api.contracts)
	mux.HandleFunc("/api/contracts/", api.contractByID)
	mux.HandleFunc("/api/settlements", api.settlements)
	mux.HandleFunc("/api/settlements/", api.settlementByID)
	mux.HandleFunc("/api/dashboard", api.dashboard)
	dist := env("FRONTEND_DIST", "../frontend/dist")
	fs := http.FileServer(http.Dir(dist))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		p := filepath.Join(dist, filepath.Clean(r.URL.Path))
		if info, e := os.Stat(p); e == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
	})
	addr := ":" + port
	log.Printf("商务管理系统已启动：http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(r *http.Request, v any) error {
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v)
}
func validate(c *Contract) error {
	if c.Type != "owner" && c.Type != "subcontract" {
		return errors.New("合同类型无效")
	}
	if strings.TrimSpace(c.Number) == "" || strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Counterparty) == "" {
		return errors.New("合同编号、名称和相对方不能为空")
	}
	if c.Amount < 0 || c.SettledAmount < 0 || c.PaidAmount < 0 {
		return errors.New("金额不能为负数")
	}
	return nil
}
func validateSettlement(s *Settlement) error {
	if s.Type != "owner" && s.Type != "subcontract" {
		return errors.New("结算类型无效")
	}
	if strings.TrimSpace(s.Number) == "" || strings.TrimSpace(s.ContractName) == "" || strings.TrimSpace(s.Counterparty) == "" {
		return errors.New("结算编号、合同名称和相对方不能为空")
	}
	if s.ReportedAmount < 0 || s.ApprovedAmount < 0 || s.PaidAmount < 0 {
		return errors.New("金额不能为负数")
	}
	return nil
}
func (a *API) contracts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		a.store.mu.RLock()
		items := append([]Contract(nil), a.store.Contracts...)
		a.store.mu.RUnlock()
		typ := r.URL.Query().Get("type")
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		out := items[:0]
		for _, c := range items {
			if typ != "" && c.Type != typ {
				continue
			}
			hay := strings.ToLower(c.Number + " " + c.Name + " " + c.Counterparty + " " + c.Manager)
			if q != "" && !strings.Contains(hay, q) {
				continue
			}
			out = append(out, c)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
		writeJSON(w, 200, out)
	case "POST":
		var c Contract
		if err := decode(r, &c); err != nil {
			writeJSON(w, 400, map[string]string{"error": "请求内容格式错误"})
			return
		}
		if err := validate(&c); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		a.store.mu.Lock()
		defer a.store.mu.Unlock()
		c.ID = a.store.NextID
		a.store.NextID++
		now := time.Now().Format(time.RFC3339)
		c.CreatedAt = now
		c.UpdatedAt = now
		a.store.Contracts = append(a.store.Contracts, c)
		if err := a.store.save(); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 201, c)
	default:
		w.WriteHeader(405)
	}
}
func (a *API) contractByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/contracts/"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a.store.mu.Lock()
	defer a.store.mu.Unlock()
	idx := -1
	for i, c := range a.store.Contracts {
		if c.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		writeJSON(w, 404, map[string]string{"error": "合同不存在"})
		return
	}
	switch r.Method {
	case "PUT":
		var c Contract
		if err := decode(r, &c); err != nil {
			writeJSON(w, 400, map[string]string{"error": "请求内容格式错误"})
			return
		}
		if err := validate(&c); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		c.ID = id
		c.CreatedAt = a.store.Contracts[idx].CreatedAt
		c.UpdatedAt = time.Now().Format(time.RFC3339)
		a.store.Contracts[idx] = c
		_ = a.store.save()
		writeJSON(w, 200, c)
	case "DELETE":
		a.store.Contracts = append(a.store.Contracts[:idx], a.store.Contracts[idx+1:]...)
		_ = a.store.save()
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}
func (a *API) settlements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		a.store.mu.RLock()
		items := append([]Settlement(nil), a.store.Settlements...)
		a.store.mu.RUnlock()
		typ := r.URL.Query().Get("type")
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		out := items[:0]
		for _, s := range items {
			if typ != "" && s.Type != typ {
				continue
			}
			hay := strings.ToLower(s.Number + " " + s.ContractNumber + " " + s.ContractName + " " + s.Counterparty + " " + s.Manager)
			if q != "" && !strings.Contains(hay, q) {
				continue
			}
			out = append(out, s)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
		writeJSON(w, 200, out)
	case "POST":
		var s Settlement
		if err := decode(r, &s); err != nil {
			writeJSON(w, 400, map[string]string{"error": "请求内容格式错误"})
			return
		}
		if err := validateSettlement(&s); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		a.store.mu.Lock()
		defer a.store.mu.Unlock()
		s.ID = a.store.NextSettlementID
		a.store.NextSettlementID++
		now := time.Now().Format(time.RFC3339)
		s.CreatedAt = now
		s.UpdatedAt = now
		a.store.Settlements = append(a.store.Settlements, s)
		if err := a.store.save(); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 201, s)
	default:
		w.WriteHeader(405)
	}
}
func (a *API) settlementByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/settlements/"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a.store.mu.Lock()
	defer a.store.mu.Unlock()
	idx := -1
	for i, s := range a.store.Settlements {
		if s.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		writeJSON(w, 404, map[string]string{"error": "结算记录不存在"})
		return
	}
	switch r.Method {
	case "PUT":
		var s Settlement
		if err := decode(r, &s); err != nil {
			writeJSON(w, 400, map[string]string{"error": "请求内容格式错误"})
			return
		}
		if err := validateSettlement(&s); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		s.ID = id
		s.CreatedAt = a.store.Settlements[idx].CreatedAt
		s.UpdatedAt = time.Now().Format(time.RFC3339)
		a.store.Settlements[idx] = s
		_ = a.store.save()
		writeJSON(w, 200, s)
	case "DELETE":
		a.store.Settlements = append(a.store.Settlements[:idx], a.store.Settlements[idx+1:]...)
		_ = a.store.save()
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}
func (a *API) dashboard(w http.ResponseWriter, r *http.Request) {
	a.store.mu.RLock()
	defer a.store.mu.RUnlock()
	res := map[string]any{"owner": map[string]float64{"count": 0, "amount": 0, "settled": 0, "paid": 0}, "subcontract": map[string]float64{"count": 0, "amount": 0, "settled": 0, "paid": 0}, "status": map[string]int{}}
	for _, c := range a.store.Contracts {
		m := res[c.Type].(map[string]float64)
		m["count"]++
		m["amount"] += c.Amount
		m["settled"] += c.SettledAmount
		m["paid"] += c.PaidAmount
		res["status"].(map[string]int)[c.Status]++
	}
	writeJSON(w, 200, res)
}
