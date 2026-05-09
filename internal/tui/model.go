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
	zone "github.com/lrstanley/bubblezone"

	"github.com/kyc-rip/cli/internal/api"
)

// zone IDs — string constants used to mark interactive regions and
// hit-test mouse clicks against them. Each session has its own
// zone manager via NewModel() so concurrent SSH sessions don't collide.
const (
	zTabSwap   = "tab-swap"
	zTabTrack  = "tab-track"
	zTabAbout  = "tab-about"
	zButton    = "button-primary"
	zAssetRow  = "asset-row-" // suffixed with index 0..8
	zAddressOK = "addr-ok"
)

// --- tabs ---

type tab int

const (
	tabSwap tab = iota
	tabTrack
	tabAbout
)

// --- swap wizard states ---

type swapState int

const (
	stPickFrom swapState = iota
	stPickTo
	stAmount
	stAddress
	stMemo // shown only when destination needs memo
	stQuoting
	stQuoted
	stCreating
	stOrdered
	stError
)

const (
	pollInterval = 5 * time.Second
	apiTimeout   = 12 * time.Second
)

// --- config ---

type Config struct {
	Client      *api.Client
	Fingerprint string
	Username    string // SSH session username for the @swap header tag

	// InitialWidth / InitialHeight let SSH hosts seed dimensions before
	// the first WindowSizeMsg arrives, so the alt-screen flip on
	// connection-start renders the form immediately instead of a void.
	InitialWidth  int
	InitialHeight int
}

// --- model ---

type Model struct {
	cfg Config

	zm *zone.Manager // per-session zone manager for mouse hit-testing

	width, height int
	tab           tab

	// wizard state
	state   swapState
	from    string // "BTC" or "USDT-TRC20"
	to      string
	amtIn   textinput.Model
	addrIn  textinput.Model
	memoIn  textinput.Model
	quote   *api.Estimate
	trade   *api.Trade
	swapErr string
	pollOn  bool

	// track tab
	trackIn    textinput.Model
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
		ti.Prompt = ""
		return ti
	}
	m := Model{
		cfg:     cfg,
		zm:      zone.New(),
		tab:     tabSwap,
		state:   stPickFrom,
		amtIn:   mk("e.g. 0.01", 24),
		addrIn:  mk("destination wallet address", 60),
		memoIn:  mk("optional memo / dest tag", 30),
		trackIn: mk("paste trade id", 40),
	}
	if cfg.InitialWidth > 0 && cfg.InitialHeight > 0 {
		m.width = cfg.InitialWidth
		m.height = cfg.InitialHeight
	}
	return m
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

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
	t       *api.Trade
	err     error
	isTrack bool
}
type tickMsg time.Time

// --- commands ---

func (m Model) cmdEstimate() tea.Cmd {
	cli := m.cfg.Client
	from, fromNet := splitTickerNet(m.from)
	to, toNet := splitTickerNet(m.to)
	amt, _ := strconv.ParseFloat(strings.TrimSpace(m.amtIn.Value()), 64)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		q, err := cli.Estimate(ctx, from, fromNet, to, toNet, amt)
		return estimateDoneMsg{q, err}
	}
}

func (m Model) cmdCreate() tea.Cmd {
	cli := m.cfg.Client
	from, fromNet := splitTickerNet(m.from)
	to, toNet := splitTickerNet(m.to)
	amt, _ := strconv.ParseFloat(strings.TrimSpace(m.amtIn.Value()), 64)
	addr := strings.TrimSpace(m.addrIn.Value())
	memo := strings.TrimSpace(m.memoIn.Value())

	provider := m.quote.Provider
	engine := m.quote.Engine
	var hq any
	if len(m.quote.Routes) > 0 {
		provider = m.quote.Routes[0].Provider
		engine = m.quote.Routes[0].Engine
		hq = m.quote.Routes[0].HoudiniQuote
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
		return statusDoneMsg{t, err, isTrack}
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

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		// Tab clicks (any state)
		if m.zm.Get(zTabSwap).InBounds(msg) {
			m.tab = tabSwap
			return m, nil
		}
		if m.zm.Get(zTabTrack).InBounds(msg) {
			m.tab = tabTrack
			m.trackIn.Focus()
			return m, nil
		}
		if m.zm.Get(zTabAbout).InBounds(msg) {
			m.tab = tabAbout
			return m, nil
		}
		// Button click → synthesize Enter
		if m.zm.Get(zButton).InBounds(msg) {
			return m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		}
		// Asset row click in pickers
		if m.tab == tabSwap && (m.state == stPickFrom || m.state == stPickTo) {
			for i := 0; i < 9 && i < len(topAssets); i++ {
				if m.zm.Get(zAssetRow + strconv.Itoa(i)).InBounds(msg) {
					m.assignAsset(topAssets[i])
					return m, nil
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Always-on shortcuts
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit
		}

		// Global hotkeys (only when NOT actively typing into the active step)
		if !m.isTypingState() {
			switch msg.String() {
			case "s":
				m.tab = tabSwap
				return m, nil
			case "t":
				m.tab = tabTrack
				m.trackIn.Focus()
				return m, nil
			case "a":
				m.tab = tabAbout
				return m, nil
			}
		}

		if m.tab == tabSwap {
			return m.updateSwap(msg)
		}
		if m.tab == tabTrack {
			return m.updateTrack(msg)
		}
		// About tab — any key returns to swap
		if msg.String() == "esc" || msg.String() == "enter" {
			m.tab = tabSwap
		}
		return m, nil

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
		if msg.isTrack {
			m.trackBusy = false
			if msg.err != nil {
				m.trackErr = msg.err.Error()
			} else {
				m.trackTrade = msg.t
			}
			return m, nil
		}
		if msg.err == nil && msg.t != nil {
			m.trade = msg.t
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

// isTypingState returns true when the active step receives raw text
// input (so we don't intercept letters as global tab hotkeys).
func (m Model) isTypingState() bool {
	if m.tab == tabTrack {
		return m.trackIn.Focused()
	}
	if m.tab != tabSwap {
		return false
	}
	switch m.state {
	case stAmount, stAddress, stMemo:
		return true
	}
	return false
}

func (m Model) updateSwap(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// step back through the wizard
		switch m.state {
		case stPickTo:
			m.state = stPickFrom
		case stAmount:
			m.state = stPickTo
		case stAddress:
			m.state = stAmount
		case stMemo:
			m.state = stAddress
		case stQuoted, stError:
			// Back to last input step
			m.state = stAddress
			m.swapErr = ""
		case stOrdered:
			// reset whole wizard
			m.resetSwap()
		}
		return m, nil
	}

	switch m.state {
	case stPickFrom, stPickTo:
		return m.updatePicker(msg)
	case stAmount:
		return m.updateAmount(msg)
	case stAddress:
		return m.updateAddress(msg)
	case stMemo:
		return m.updateMemo(msg)
	case stQuoted:
		if msg.String() == "enter" {
			m.state = stCreating
			return m, m.cmdCreate()
		}
	case stError:
		if msg.String() == "enter" {
			m.swapErr = ""
			m.state = stAddress
		}
	}
	return m, nil
}

func (m *Model) resetSwap() {
	m.state = stPickFrom
	m.from = ""
	m.to = ""
	m.amtIn.SetValue("")
	m.addrIn.SetValue("")
	m.memoIn.SetValue("")
	m.quote = nil
	m.trade = nil
	m.swapErr = ""
	m.pollOn = false
}

// updatePicker handles stPickFrom / stPickTo: digits 1-9 pick from the
// numbered list; any other typing is interpreted as a free-text ticker
// followed by Enter.
func (m Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	// Digit shortcut
	if len(key) == 1 && key >= "1" && key <= "9" {
		idx := int(key[0]-'0') - 1
		if idx >= 0 && idx < len(topAssets) {
			m.assignAsset(topAssets[idx])
			return m, nil
		}
	}
	// Free-text typing: keep an input visible at the bottom.
	// Reuse amtIn as a temporary scratch input — too much state otherwise.
	switch key {
	case "enter":
		txt := strings.TrimSpace(m.pickerScratch())
		if txt == "" {
			return m, nil
		}
		m.assignAsset(strings.ToUpper(txt))
		m.setPickerScratch("")
		return m, nil
	case "backspace":
		s := m.pickerScratch()
		if len(s) > 0 {
			m.setPickerScratch(s[:len(s)-1])
		}
		return m, nil
	}
	if len(key) == 1 {
		m.setPickerScratch(m.pickerScratch() + key)
	}
	return m, nil
}

// assignAsset commits the picked ticker to the current step and advances.
func (m *Model) assignAsset(t string) {
	t = strings.ToUpper(strings.TrimSpace(t))
	if m.state == stPickFrom {
		m.from = t
		m.state = stPickTo
		return
	}
	if m.state == stPickTo {
		m.to = t
		m.state = stAmount
		m.amtIn.Focus()
	}
}

func (m Model) updateAmount(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		amt, err := strconv.ParseFloat(strings.TrimSpace(m.amtIn.Value()), 64)
		if err != nil || amt <= 0 {
			m.swapErr = "amount must be a positive number"
			return m, nil
		}
		m.amtIn.Blur()
		m.state = stAddress
		m.addrIn.Focus()
		m.swapErr = ""
		return m, nil
	}
	var cmd tea.Cmd
	m.amtIn, cmd = m.amtIn.Update(msg)
	return m, cmd
}

func (m Model) updateAddress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		addr := strings.TrimSpace(m.addrIn.Value())
		if addr == "" {
			m.swapErr = "destination address required"
			return m, nil
		}
		// Pre-flight format heuristic — catches typos before sending
		// the order to upstream engines (which may take real $$ to
		// learn the address was malformed).
		toTicker, toNet := splitTickerNet(m.to)
		if ok, hint := validateAddress(toTicker, toNet, addr); !ok {
			m.swapErr = hint
			return m, nil
		}
		m.addrIn.Blur()
		// Memo is requested upfront for everything; we always show the step
		// but blank-and-Enter skips it cleanly.
		m.state = stMemo
		m.memoIn.Focus()
		m.swapErr = ""
		return m, nil
	}
	var cmd tea.Cmd
	m.addrIn, cmd = m.addrIn.Update(msg)
	return m, cmd
}

func (m Model) updateMemo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m.memoIn.Blur()
		m.state = stQuoting
		return m, m.cmdEstimate()
	}
	var cmd tea.Cmd
	m.memoIn, cmd = m.memoIn.Update(msg)
	return m, cmd
}

func (m Model) updateTrack(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		id := strings.TrimSpace(m.trackIn.Value())
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
		m.trackIn.SetValue("")
		return m, nil
	}
	var cmd tea.Cmd
	m.trackIn, cmd = m.trackIn.Update(msg)
	return m, cmd
}

// --- picker scratch (free-text typing buffer for asset pickers) ---

func (m *Model) pickerScratch() string {
	if m.state == stPickFrom {
		return m.from // reuse field as buffer
	}
	return m.to
}
func (m *Model) setPickerScratch(s string) {
	if m.state == stPickFrom {
		m.from = s
	} else {
		m.to = s
	}
}

// --- helpers ---

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
	return strconv.FormatFloat(n, 'f', -1, 64)
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
	case tabAbout:
		body = m.renderAbout()
	}
	hint := m.renderHint()

	// Centered card layout.
	card := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	cardBox := styleCard.Render(card)

	// Stack: card centered + hint bar at bottom
	stack := lipgloss.JoinVertical(lipgloss.Center, cardBox, "", styleDim.Render(hint))
	placed := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, stack)
	// Scan registers the rendered zone bounds so MouseMsg.InBounds() works.
	return m.zm.Scan(placed)
}

func (m Model) renderHeader() string {
	user := m.cfg.Username
	if user == "" {
		user = "kyc.rip"
	}
	left := styleUser.Render(user + "@swap")
	tabs := []string{
		m.zm.Mark(zTabSwap, tabRender("Swap", m.tab == tabSwap)),
		m.zm.Mark(zTabTrack, tabRender("Track", m.tab == tabTrack)),
		m.zm.Mark(zTabAbout, tabRender("About", m.tab == tabAbout)),
	}
	right := strings.Join(tabs, "  ")
	// Spacer flex
	spacerWidth := cardInnerWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	spacer := strings.Repeat(" ", spacerWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
}

// tabRender wraps each tab in its own zone so mouse clicks resolve to it.
// We don't have access to the zone manager here so the caller wraps after.
func tabRender(name string, active bool) string {
	if active {
		return styleTabActive.Render(name)
	}
	return styleTabIdle.Render(name)
}

func (m Model) renderSwap() string {
	switch m.state {
	case stPickFrom:
		return m.renderPicker("Sending", m.from)
	case stPickTo:
		return m.renderPicker("Receiving", m.to)
	case stAmount:
		return m.renderAmount()
	case stAddress:
		return m.renderAddress()
	case stMemo:
		return m.renderMemo()
	case stQuoting:
		return styleDim.Render("fetching best quote across engines…")
	case stQuoted:
		return m.renderQuoted()
	case stCreating:
		return styleDim.Render("creating order…")
	case stOrdered:
		return m.renderOrdered()
	case stError:
		return styleErr.Render("error: ") + m.swapErr + "\n\n" + styleDim.Render("press enter to retry · esc to step back")
	}
	return ""
}

func (m Model) renderPicker(label, scratch string) string {
	var rows []string
	rows = append(rows, styleAccent.Render(label+":")+" "+styleDim.Render("pick a number 1-9, click a row, or type a ticker"))
	rows = append(rows, "")
	for i, a := range topAssets {
		if i >= 9 {
			break
		}
		row := fmt.Sprintf("  %s  %s", styleWarn.Render(strconv.Itoa(i+1)+"."), a)
		rows = append(rows, m.zm.Mark(zAssetRow+strconv.Itoa(i), row))
	}
	rows = append(rows, "")
	rows = append(rows, styleDim.Render("type:")+" "+styleField.Render(padInput(scratch, 28)))
	rows = append(rows, "")
	rows = append(rows, m.zm.Mark(zButton, styleButton.Render("[ Enter to confirm ]")))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderAmount() string {
	rows := []string{
		styleAccent.Render("Sending: ") + m.from,
		styleDim.Render("amount you send"),
		"",
		styleFieldActive.Render(padInput(m.amtIn.View(), 24)),
	}
	if m.swapErr != "" {
		rows = append(rows, "", styleErr.Render(m.swapErr))
	}
	rows = append(rows, "", m.zm.Mark(zButton, styleButton.Render("[ Continue ]")))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderAddress() string {
	rows := []string{
		styleAccent.Render("Receiving: ") + m.to,
		styleDim.Render("destination wallet address"),
		"",
		styleFieldActive.Render(padInput(m.addrIn.View(), 60)),
	}
	if m.swapErr != "" {
		rows = append(rows, "", styleErr.Render(m.swapErr))
	}
	rows = append(rows, "", m.zm.Mark(zButton, styleButton.Render("[ Continue ]")))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderMemo() string {
	rows := []string{
		styleAccent.Render("Memo / dest tag"),
		styleDim.Render("optional · leave blank and press Enter to skip"),
		"",
		styleFieldActive.Render(padInput(m.memoIn.View(), 30)),
		"",
		m.zm.Mark(zButton, styleButton.Render("[ Get quote ]")),
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderQuoted() string {
	q := m.quote
	if q == nil {
		return ""
	}
	from, _ := splitTickerNet(m.from)
	to, _ := splitTickerNet(m.to)
	routeName := q.Provider
	if len(q.Routes) > 0 {
		routeName = q.Routes[0].Provider
	}
	rows := []string{
		styleAccent.Render("Sending:   ") + fmtAmt(q.AmountFrom) + " " + from,
		styleAccent.Render("Receiving: ") + styleOk.Render("~"+fmtAmt(q.AmountTo)+" "+to),
		"",
		styleDim.Render(fmt.Sprintf("1 %s = %s %s", from, fmtAmt(q.Rate), to)),
		styleDim.Render(fmt.Sprintf("via %s · ETA ~%dm · KYC %s", routeName, q.ETA, q.KYCRating)),
		"",
		m.zm.Mark(zButton, styleButton.Render("[ Confirm — create order ]")),
		"",
		styleDim.Render("enter confirm · esc back"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderOrdered() string {
	t := m.trade
	if t == nil {
		return ""
	}
	qr := renderQR(t.DepositAddress)
	left := []string{
		styleAccent.Render("Order ") + t.ID,
		styleAccent.Render("Status: ") + styleOk.Render(strings.ToUpper(t.Status)),
		"",
		styleDim.Render("Send"),
		styleOk.Render(fmt.Sprintf("%s %s", fmtAmt(t.FromAmount), strings.ToUpper(t.FromTicker))),
		"",
		styleDim.Render("To deposit address"),
		styleOk.Render(t.DepositAddress),
	}
	if t.DepositMemo != "" {
		left = append(left, "", styleDim.Render("Memo (REQUIRED)"), styleErr.Render(t.DepositMemo))
	}
	left = append(left, "",
		styleDim.Render(fmt.Sprintf("Receive ~%s %s → %s", fmtAmt(t.ToAmount), strings.ToUpper(t.ToTicker), t.AddressUser)),
		"",
		styleDim.Render("auto-refresh every 5s · esc reset"),
	)
	leftBlock := lipgloss.JoinVertical(lipgloss.Left, left...)
	if qr == "" {
		return leftBlock
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, "  ", qr)
}

func (m Model) renderTrack() string {
	rows := []string{
		styleAccent.Render("Track"),
		styleDim.Render("paste a trade id and press Enter"),
		"",
		styleFieldActive.Render(padInput(m.trackIn.View(), 40)),
	}
	if m.trackBusy {
		rows = append(rows, "", styleDim.Render("looking up…"))
	} else if m.trackErr != "" {
		rows = append(rows, "", styleErr.Render("error: "+m.trackErr))
	} else if m.trackTrade != nil {
		t := m.trackTrade
		rows = append(rows, "",
			styleAccent.Render("Status: ")+styleOk.Render(strings.ToUpper(t.Status)),
			styleDim.Render(fmt.Sprintf("send %s %s", fmtAmt(t.FromAmount), strings.ToUpper(t.FromTicker))),
			styleDim.Render(fmt.Sprintf("recv %s %s → %s", fmtAmt(t.ToAmount), strings.ToUpper(t.ToTicker), t.AddressUser)),
			styleDim.Render("txIn:  "+t.TxIn),
			styleDim.Render("txOut: "+t.TxOut),
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderAbout() string {
	fp := m.cfg.Fingerprint
	if fp == "" {
		fp = "(local CLI · no host key)"
	}
	rows := []string{
		styleAccent.Render("kyc.rip · terminal-only swap"),
		"",
		styleDim.Render("Privacy-first crypto swap aggregator,"),
		styleDim.Render("served over SSH. No JS, no cookies,"),
		styleDim.Render("no browser fingerprint."),
		"",
		styleAccent.Render("Channels"),
		"  clearnet  ssh swap.kyc.rip",
		"  https     https://swap.kyc.rip  (landing only)",
		"  tor       torsocks ssh ozz6kgrbp6epsxhrid456",
		"            udvwj3vzecb4f7jz5orxcrpxn4f2bejuyid.onion",
		"  i2p       ssh -o ProxyCommand='nc -X 5 -x 127.0.0.1:4447 %h %p' \\",
		"               r4ziaqaec7w73x7ltpz5pi5kswclgjdw6",
		"               ioyz25mbtrisprneqhq.b32.i2p",
		"",
		styleAccent.Render("Verify host key before connecting"),
		"  " + fp,
		"",
		styleAccent.Render("Source"),
		"  github.com/kyc-rip/cli",
		"",
		styleDim.Render("press s · t · ctrl+c"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderHint() string {
	switch m.tab {
	case tabSwap:
		switch m.state {
		case stPickFrom, stPickTo:
			return "1-9 pick · type ticker · enter confirm · s swap · t track · a about · ctrl+c quit"
		case stAmount, stAddress, stMemo:
			return "type · enter continue · esc back · ctrl+c quit"
		case stQuoted:
			return "enter confirm · esc back · t track · ctrl+c quit"
		case stOrdered:
			return "esc reset · t track · ctrl+c quit"
		}
	case tabTrack:
		return "type id · enter lookup · esc clear · s swap · a about · ctrl+c quit"
	case tabAbout:
		return "esc/enter back · s swap · t track · ctrl+c quit"
	}
	return "ctrl+c quit"
}

// padInput right-pads s to width w (preserving lipgloss-rendered width
// where possible) so input fields don't reflow when text is empty.
func padInput(s string, w int) string {
	cur := lipgloss.Width(s)
	if cur >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cur)
}
