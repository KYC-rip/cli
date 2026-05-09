package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/xbtoshi/sshwap/internal/api"
)

type tab int

const (
	tabSwap tab = iota
	tabTrack
)

type swapState int

const (
	stInput swapState = iota
	stQuoting
	stQuoted
	stCreating
	stOrdered
	stError
)

const (
	fldFrom = iota
	fldTo
	fldAmount
	fldAddress
	fldMemo
	fldButton
	numFields
)

const (
	pollInterval = 5 * time.Second
	apiTimeout   = 12 * time.Second
)

type Config struct {
	Client      *api.Client
	Fingerprint string
	HostBanner  string
}

type Model struct {
	cfg Config

	width, height int
	tab           tab

	// swap form
	state    swapState
	field    int
	from     textinput.Model
	to       textinput.Model
	amount   textinput.Model
	address  textinput.Model
	memo     textinput.Model
	quote    *api.Estimate
	trade    *api.Trade
	swapErr  string
	pollOn   bool
	pollTick time.Time

	// track form
	trackID    textinput.Model
	trackTrade *api.Trade
	trackErr   string
	trackBusy  bool
}

func New(cfg Config) Model {
	mk := func(ph string, w int) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = ph
		ti.CharLimit = 128
		ti.Width = w
		ti.Prompt = "▎ "
		return ti
	}
	from := mk("e.g. BTC, USDT-TRC20", 28)
	from.Focus()
	m := Model{
		cfg:     cfg,
		tab:     tabSwap,
		state:   stInput,
		field:   fldFrom,
		from:    from,
		to:      mk("e.g. XMR, ETH-ERC20", 28),
		amount:  mk("e.g. 0.01", 18),
		address: mk("destination wallet address", 60),
		memo:    mk("optional memo / dest tag", 30),
		trackID: mk("paste trade id", 40),
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// --- messages ---

type estimateDoneMsg struct {
	q   *api.Estimate
	err error
}
type tradeDoneMsg struct {
	t   *api.Trade
	err error
}
type statusDoneMsg struct {
	t   *api.Trade
	err error
}
type tickMsg time.Time

// --- commands ---

func (m Model) cmdEstimate(from, fromNet, to, toNet string, amt float64) tea.Cmd {
	cli := m.cfg.Client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		q, err := cli.Estimate(ctx, from, fromNet, to, toNet, amt)
		return estimateDoneMsg{q, err}
	}
}

func (m Model) cmdCreate(req api.CreateReq) tea.Cmd {
	cli := m.cfg.Client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		t, err := cli.Create(ctx, req)
		return tradeDoneMsg{t, err}
	}
}

func (m Model) cmdStatus(id string, isTrack bool) tea.Cmd {
	cli := m.cfg.Client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		t, err := cli.Status(ctx, id)
		if isTrack {
			return statusDoneMsg{t, err}
		}
		return statusDoneMsg{t, err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// --- update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		// Always-on shortcuts
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit
		case "tab":
			if m.tab == tabSwap {
				m.tab = tabTrack
				m.blurAll()
				m.trackID.Focus()
			} else {
				m.tab = tabSwap
				m.trackID.Blur()
				m.focusField()
			}
			return m, nil
		}
		if m.tab == tabSwap {
			return m.updateSwap(msg)
		}
		return m.updateTrack(msg)

	case estimateDoneMsg:
		if msg.err != nil {
			m.state = stError
			m.swapErr = msg.err.Error()
			return m, nil
		}
		m.quote = msg.q
		m.state = stQuoted
		return m, nil

	case tradeDoneMsg:
		if msg.err != nil {
			m.state = stError
			m.swapErr = msg.err.Error()
			return m, nil
		}
		m.trade = msg.t
		m.state = stOrdered
		m.pollOn = true
		return m, tickCmd()

	case statusDoneMsg:
		if msg.err == nil && msg.t != nil {
			if m.tab == tabTrack {
				m.trackTrade = msg.t
				m.trackBusy = false
			} else {
				m.trade = msg.t
			}
		} else if msg.err != nil && m.tab == tabTrack {
			m.trackErr = msg.err.Error()
			m.trackBusy = false
		}
		return m, nil

	case tickMsg:
		if m.pollOn && m.trade != nil && m.trade.ID != "" && !isTerminal(m.trade.Status) {
			return m, tea.Batch(m.cmdStatus(m.trade.ID, false), tickCmd())
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) blurAll() {
	m.from.Blur()
	m.to.Blur()
	m.amount.Blur()
	m.address.Blur()
	m.memo.Blur()
}

func (m *Model) focusField() {
	m.blurAll()
	switch m.field {
	case fldFrom:
		m.from.Focus()
	case fldTo:
		m.to.Focus()
	case fldAmount:
		m.amount.Focus()
	case fldAddress:
		m.address.Focus()
	case fldMemo:
		m.memo.Focus()
	}
}

func (m Model) updateSwap(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.state == stOrdered || m.state == stError || m.state == stQuoted {
			m.resetSwap()
			return m, nil
		}
	case "up", "shift+tab":
		if m.state == stInput {
			m.field = (m.field - 1 + numFields) % numFields
			m.focusField()
			return m, nil
		}
	case "down":
		if m.state == stInput {
			m.field = (m.field + 1) % numFields
			m.focusField()
			return m, nil
		}
	case "enter":
		switch m.state {
		case stInput:
			if m.field == fldButton || m.field == fldMemo {
				return m.submitQuote()
			}
			m.field = (m.field + 1) % numFields
			m.focusField()
			return m, nil
		case stQuoted:
			return m.submitCreate()
		}
	}

	// pass through to active textinput
	if m.state == stInput {
		var cmd tea.Cmd
		switch m.field {
		case fldFrom:
			m.from, cmd = m.from.Update(msg)
		case fldTo:
			m.to, cmd = m.to.Update(msg)
		case fldAmount:
			m.amount, cmd = m.amount.Update(msg)
		case fldAddress:
			m.address, cmd = m.address.Update(msg)
		case fldMemo:
			m.memo, cmd = m.memo.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

func (m *Model) resetSwap() {
	m.state = stInput
	m.quote = nil
	m.trade = nil
	m.swapErr = ""
	m.pollOn = false
	m.field = fldFrom
	m.focusField()
}

func (m Model) submitQuote() (tea.Model, tea.Cmd) {
	from, fromNet := splitTickerNet(strings.TrimSpace(m.from.Value()))
	to, toNet := splitTickerNet(strings.TrimSpace(m.to.Value()))
	amtStr := strings.TrimSpace(m.amount.Value())
	addr := strings.TrimSpace(m.address.Value())
	if from == "" || to == "" || amtStr == "" || addr == "" {
		m.swapErr = "from, to, amount and address are required"
		m.state = stError
		return m, nil
	}
	amt, err := strconv.ParseFloat(amtStr, 64)
	if err != nil || amt <= 0 {
		m.swapErr = "amount must be a positive number"
		m.state = stError
		return m, nil
	}
	m.state = stQuoting
	m.swapErr = ""
	return m, m.cmdEstimate(from, fromNet, to, toNet, amt)
}

func (m Model) submitCreate() (tea.Model, tea.Cmd) {
	if m.quote == nil {
		return m, nil
	}
	from, fromNet := splitTickerNet(strings.TrimSpace(m.from.Value()))
	to, toNet := splitTickerNet(strings.TrimSpace(m.to.Value()))
	amt, _ := strconv.ParseFloat(strings.TrimSpace(m.amount.Value()), 64)
	addr := strings.TrimSpace(m.address.Value())
	memo := strings.TrimSpace(m.memo.Value())

	// Pick best route — use the top one returned by the aggregator.
	// (POC: aggregator already sorts; future: let user pick.)
	var route *api.Route
	if len(m.quote.Routes) > 0 {
		route = &m.quote.Routes[0]
	}
	provider := m.quote.Provider
	engine := m.quote.Engine
	var hq any
	if route != nil {
		provider = route.Provider
		engine = route.Engine
		hq = route.HoudiniQuote
	}

	req := api.CreateReq{
		Provider:     provider,
		Engine:       engine,
		FromCurrency: from,
		ToCurrency:   to,
		FromNetwork:  fromNet,
		ToNetwork:    toNet,
		AmountFrom:   amt,
		AddressTo:    addr,
		AddressMemo:  memo,
		HoudiniQuote: hq,
	}
	m.state = stCreating
	return m, m.cmdCreate(req)
}

func (m Model) updateTrack(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		id := strings.TrimSpace(m.trackID.Value())
		if id == "" {
			return m, nil
		}
		m.trackBusy = true
		m.trackErr = ""
		m.trackTrade = nil
		return m, m.cmdStatus(id, true)
	case "esc":
		m.trackTrade = nil
		m.trackErr = ""
		m.trackID.SetValue("")
		return m, nil
	}
	var cmd tea.Cmd
	m.trackID, cmd = m.trackID.Update(msg)
	return m, cmd
}

// --- helpers ---

// splitTickerNet accepts "BTC", "USDT-TRC20", "USDT/TRC20".
// Returns ticker, network (network "" means default Mainnet).
func splitTickerNet(in string) (string, string) {
	in = strings.ToUpper(strings.TrimSpace(in))
	for _, sep := range []string{"-", "/", ":"} {
		if i := strings.Index(in, sep); i > 0 {
			return in[:i], in[i+1:]
		}
	}
	return in, ""
}

func isTerminal(status string) bool {
	switch strings.ToUpper(status) {
	case "FINISHED", "FAILED", "EXPIRED", "REFUNDED":
		return true
	}
	return false
}

func fmtAmt(n float64) string {
	if n == 0 {
		return "—"
	}
	s := strconv.FormatFloat(n, 'f', -1, 64)
	return s
}

// --- view ---

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	header := m.renderHeader()
	var body string
	switch m.tab {
	case tabSwap:
		body = m.renderSwap()
	case tabTrack:
		body = m.renderTrack()
	}
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
}

func (m Model) renderHeader() string {
	tabSwap := "Swap"
	tabTrack := "Track"
	if m.tab == 0 {
		tabSwap = styleTabActive.Render(tabSwap)
		tabTrack = styleTabIdle.Render(tabTrack)
	} else {
		tabSwap = styleTabIdle.Render(tabSwap)
		tabTrack = styleTabActive.Render(tabTrack)
	}
	title := styleTitle.Render("kyc.rip — swap")
	return lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		"   ",
		tabSwap, "  ", tabTrack,
	)
}

func (m Model) renderSwap() string {
	switch m.state {
	case stOrdered:
		return m.renderOrdered()
	case stQuoted:
		return m.renderQuoted()
	case stQuoting:
		return styleDim.Render("fetching best quote across engines…")
	case stCreating:
		return styleDim.Render("creating order…")
	case stError:
		return styleErr.Render("error: ") + m.swapErr + "\n\n" + styleDim.Render("press esc to start over")
	}
	return m.renderForm()
}

func (m Model) renderForm() string {
	field := func(label string, ti textinput.Model, idx int) string {
		st := styleField
		if m.field == idx {
			st = styleFieldActive
		}
		return styleDim.Render(label) + "\n" + st.Render(ti.View())
	}
	btn := styleButtonIdle.Render("[ Get quote ]")
	if m.field == fldButton {
		btn = styleButton.Render("[ Get quote ]")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		field("From asset (TICKER or TICKER-NET)", m.from, fldFrom),
		field("To asset", m.to, fldTo),
		field("Amount (you send)", m.amount, fldAmount),
		field("Destination address", m.address, fldAddress),
		field("Memo / dest tag (optional)", m.memo, fldMemo),
		"",
		btn,
		"",
		styleDim.Render("↑/↓ navigate · enter advance / submit · tab switch tab · ctrl+c quit"),
	)
}

func (m Model) renderQuoted() string {
	q := m.quote
	if q == nil {
		return ""
	}
	var routeName string
	if len(q.Routes) > 0 {
		routeName = q.Routes[0].Provider
	} else {
		routeName = q.Provider
	}
	rate := fmt.Sprintf("1 %s ≈ %s %s", strings.ToUpper(strings.Split(strings.TrimSpace(m.from.Value()), "-")[0]),
		strconv.FormatFloat(q.Rate, 'f', -1, 64),
		strings.ToUpper(strings.Split(strings.TrimSpace(m.to.Value()), "-")[0]))
	return lipgloss.JoinVertical(lipgloss.Left,
		styleTitle.Render("Quote"),
		fmt.Sprintf("Route:    %s (%s)", routeName, q.Engine),
		fmt.Sprintf("You send: %s", fmtAmt(q.AmountFrom)),
		fmt.Sprintf("You get:  %s   (ETA ~%dm)", styleOk.Render(fmtAmt(q.AmountTo)), q.ETA),
		fmt.Sprintf("Rate:     %s", rate),
		fmt.Sprintf("KYC:      %s", q.KYCRating),
		"",
		styleButton.Render("[ Confirm — create order ]"),
		"",
		styleDim.Render("enter confirm · esc cancel · ctrl+c quit"),
	)
}

func (m Model) renderOrdered() string {
	t := m.trade
	if t == nil {
		return ""
	}
	qr := renderQR(t.DepositAddress)
	left := lipgloss.JoinVertical(lipgloss.Left,
		styleTitle.Render("Order created"),
		fmt.Sprintf("ID:       %s", t.ID),
		fmt.Sprintf("Status:   %s", styleOk.Render(strings.ToUpper(t.Status))),
		fmt.Sprintf("Send:     %s %s", fmtAmt(t.FromAmount), strings.ToUpper(t.FromTicker)),
		fmt.Sprintf("To addr:  %s", styleOk.Render(t.DepositAddress)),
		func() string {
			if t.DepositMemo != "" {
				return fmt.Sprintf("Memo:     %s", styleErr.Render(t.DepositMemo))
			}
			return ""
		}(),
		fmt.Sprintf("Receive:  %s %s → %s", fmtAmt(t.ToAmount), strings.ToUpper(t.ToTicker), t.AddressUser),
		"",
		styleDim.Render("auto-refreshing every 5s · esc reset · ctrl+c quit"),
	)
	if qr == "" {
		return left
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", qr)
}

func (m Model) renderTrack() string {
	out := lipgloss.JoinVertical(lipgloss.Left,
		styleDim.Render("Trade ID"),
		styleFieldActive.Render(m.trackID.View()),
		"",
		styleDim.Render("enter to look up · esc to clear"),
	)
	if m.trackBusy {
		return out + "\n\n" + styleDim.Render("looking up…")
	}
	if m.trackErr != "" {
		return out + "\n\n" + styleErr.Render("error: "+m.trackErr)
	}
	if m.trackTrade != nil {
		t := m.trackTrade
		out += "\n\n" + lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("Status:   %s", styleOk.Render(strings.ToUpper(t.Status))),
			fmt.Sprintf("Send:     %s %s", fmtAmt(t.FromAmount), strings.ToUpper(t.FromTicker)),
			fmt.Sprintf("Receive:  %s %s → %s", fmtAmt(t.ToAmount), strings.ToUpper(t.ToTicker), t.AddressUser),
			fmt.Sprintf("TxIn:     %s", t.TxIn),
			fmt.Sprintf("TxOut:    %s", t.TxOut),
		)
	}
	return out
}

func (m Model) renderFooter() string {
	fp := m.cfg.Fingerprint
	if fp == "" {
		fp = "(host key fingerprint not configured)"
	}
	line := fmt.Sprintf("kyc.rip · host key %s", fp)
	return styleFooter.Width(m.width).Render(line)
}
