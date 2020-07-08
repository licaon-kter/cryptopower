package ui

import (
	"image"

	"github.com/raedahgroup/godcr/ui/values"

	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/raedahgroup/godcr/ui/decredmaterial"
	"github.com/raedahgroup/godcr/wallet"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type pageIcons struct {
	contentAdd, contentClear, contentCreate, navigationCheck,
	contentSend, contentAddBox, contentRemove, toggleRadioButtonUnchecked,
	actionCheckCircle, contentCopy, actionInfo, navigationMore,
	navigationArrowBack, verifyAction, actionDelete, actionLock,
	communicationComment, editorModeEdit, actionBackup, actionCheck,
	actionSwapVert, navigationCancel, notificationSync, imageBrightness1 *decredmaterial.Icon

	overviewIcon, walletIcon, receiveIcon, transactionIcon, sendIcon, syncingIcon image.Image
}

type pageCommon struct {
	wallet          *wallet.Wallet
	info            *wallet.MultiWalletInfo
	selectedWallet  *int
	selectedAccount *int
	gtx             *layout.Context
	theme           *decredmaterial.Theme
	icons           pageIcons
	page            *string
	navTab          *decredmaterial.Tabs
	walletsTab      *decredmaterial.Tabs
	accountsTab     *decredmaterial.Tabs
	errorChannels   map[string]chan error
	keyEvents       chan *key.Event
	states          *states
}

func (win *Window) addPages(decredIcons map[string]image.Image) {
	icons := pageIcons{
		contentAdd:                 mustIcon(decredmaterial.NewIcon(icons.ContentAdd)),
		contentClear:               mustIcon(decredmaterial.NewIcon(icons.ContentClear)),
		contentCreate:              mustIcon(decredmaterial.NewIcon(icons.ContentCreate)),
		navigationCheck:            mustIcon(decredmaterial.NewIcon(icons.NavigationCheck)),
		contentSend:                mustIcon(decredmaterial.NewIcon(icons.ContentSend)),
		contentAddBox:              mustIcon(decredmaterial.NewIcon(icons.ContentAddBox)),
		contentRemove:              mustIcon(decredmaterial.NewIcon(icons.ContentRemove)),
		toggleRadioButtonUnchecked: mustIcon(decredmaterial.NewIcon(icons.ToggleRadioButtonUnchecked)),
		actionCheckCircle:          mustIcon(decredmaterial.NewIcon(icons.ActionCheckCircle)),
		navigationArrowBack:        mustIcon(decredmaterial.NewIcon(icons.NavigationArrowBack)),
		contentCopy:                mustIcon(decredmaterial.NewIcon(icons.NavigationMoreVert)),
		actionInfo:                 mustIcon(decredmaterial.NewIcon(icons.ActionInfo)),
		navigationMore:             mustIcon(decredmaterial.NewIcon(icons.NavigationMoreVert)),
		actionDelete:               mustIcon(decredmaterial.NewIcon(icons.ActionDelete)),
		communicationComment:       mustIcon(decredmaterial.NewIcon(icons.CommunicationComment)),
		verifyAction:               mustIcon(decredmaterial.NewIcon(icons.ActionVerifiedUser)),
		editorModeEdit:             mustIcon(decredmaterial.NewIcon(icons.EditorModeEdit)),
		actionLock:                 mustIcon(decredmaterial.NewIcon(icons.ActionLock)),
		actionBackup:               mustIcon(decredmaterial.NewIcon(icons.ActionSettingsBackupRestore)),
		actionCheck:                mustIcon(decredmaterial.NewIcon(icons.ActionCheckCircle)),
		actionSwapVert:             mustIcon(decredmaterial.NewIcon(icons.ActionSwapVert)),
		navigationCancel:           mustIcon(decredmaterial.NewIcon(icons.NavigationCancel)),
		notificationSync:           mustIcon(decredmaterial.NewIcon(icons.NotificationSync)),
		imageBrightness1:           mustIcon(decredmaterial.NewIcon(icons.ImageBrightness1)),
		overviewIcon:               decredIcons["overview"],
		walletIcon:                 decredIcons["wallet"],
		receiveIcon:                decredIcons["receive"],
		transactionIcon:            decredIcons["transaction"],
		sendIcon:                   decredIcons["send"],
		syncingIcon:                decredIcons["syncing"],
	}

	tabs := decredmaterial.NewTabs()
	tabs.SetTabs([]decredmaterial.TabItem{
		{
			Label: win.theme.Body1("Overview"),
			Icon:  icons.overviewIcon,
		},
		{
			Label: win.theme.Body1("Wallets"),
			Icon:  icons.walletIcon,
		},
		{
			Label: win.theme.Body1("Send"),
			Icon:  icons.sendIcon,
		},
		{
			Label: win.theme.Body1("Receive"),
			Icon:  icons.receiveIcon,
		},
		{
			Label: win.theme.Body1("Transactions"),
			Icon:  icons.transactionIcon,
		},
	})

	accountsTab := decredmaterial.NewTabs()
	accountsTab.Position = decredmaterial.Top
	accountsTab.Separator = false
	common := pageCommon{
		wallet:          win.wallet,
		info:            win.walletInfo,
		selectedWallet:  &win.selected,
		selectedAccount: &win.selectedAccount,
		gtx:             win.gtx,
		theme:           win.theme,
		icons:           icons,
		page:            &win.current,
		navTab:          tabs,
		walletsTab:      decredmaterial.NewTabs(),
		accountsTab:     accountsTab,
		errorChannels: map[string]chan error{
			PageSignMessage:    make(chan error),
			PageCreateRestore:  make(chan error),
			PageWallet:         make(chan error),
			PageWalletAccounts: make(chan error),
		},
		keyEvents: win.keyEvents,
		states:    &win.states,
	}

	win.pages = make(map[string]layout.Widget)
	win.pages[PageWallet] = win.WalletPage(common)
	win.pages[PageOverview] = win.OverviewPage(common)
	win.pages[PageTransactions] = win.TransactionsPage(common)
	win.pages[PageCreateRestore] = win.CreateRestorePage(common)
	win.pages[PageReceive] = win.ReceivePage(common)
	win.pages[PageSend] = win.SendPage(common)
	win.pages[PageTransactionDetails] = win.TransactionPage(common)
	win.pages[PageSignMessage] = win.SignMessagePage(common)
	win.pages[PageVerifyMessage] = win.VerifyMessagePage(common)
	win.pages[PageWalletPassphrase] = win.WalletPassphrasePage(common)
	win.pages[PageWalletAccounts] = win.WalletAccountPage(common)
	win.pages[PageSeedBackup] = win.BackupPage(common)
}

func (page pageCommon) Layout(gtx *layout.Context, body layout.Widget) {
	navs := []string{PageOverview, PageWallet, PageSend, PageReceive, PageTransactions}
	toMax(gtx)
	page.navTab.Separator = true
	page.navTab.Layout(gtx, func() {
		p := values.MarginPadding10
		layout.Inset{Top: p, Left: p, Right: p}.Layout(page.gtx, func() {
			body()
		})
	})

	for range page.navTab.ChangeEvent(gtx) {
		*page.page = navs[page.navTab.Selected]
	}
}

func (page pageCommon) LayoutWithWallets(gtx *layout.Context, body layout.Widget) {
	wallets := make([]decredmaterial.TabItem, len(page.info.Wallets))
	for i := range page.info.Wallets {
		wallets[i] = decredmaterial.TabItem{
			Label: page.theme.Body1(page.info.Wallets[i].Name),
		}
	}
	page.walletsTab.SetTabs(wallets)
	page.walletsTab.Position = decredmaterial.Top
	for range page.accountsTab.ChangeEvent(gtx) {
		*page.selectedAccount = page.accountsTab.Selected
	}

	accounts := make([]decredmaterial.TabItem, len(page.info.Wallets[*page.selectedWallet].Accounts))
	for i, acct := range page.info.Wallets[*page.selectedWallet].Accounts {
		if acct.Name == "imported" {
			continue
		}
		accounts[i] = decredmaterial.TabItem{
			Label: page.theme.Body1(page.info.Wallets[*page.selectedWallet].Accounts[i].Name),
		}
	}
	page.accountsTab.SetTabs(accounts)
	for range page.accountsTab.ChangeEvent(gtx) {
		*page.selectedAccount = page.accountsTab.Selected
	}
	page.accountsTab.Separator = false

	bd := func() {
		for range page.walletsTab.ChangeEvent(gtx) {
			*page.selectedWallet = page.walletsTab.Selected
			*page.selectedAccount = 0
			page.accountsTab.Selected = 0
		}
		if *page.selectedWallet == 0 {
			page.walletsTab.Selected = *page.selectedWallet
		}
		page.walletsTab.Separator = false
		page.walletsTab.Layout(gtx, body)
	}
	page.Layout(gtx, bd)
}

func (page pageCommon) LayoutWithAccounts(gtx *layout.Context, body layout.Widget) {
	accounts := make([]decredmaterial.TabItem, len(page.info.Wallets[*page.selectedWallet].Accounts))
	for i, account := range page.info.Wallets[*page.selectedWallet].Accounts {
		if account.Name == "imported" {
			continue
		}
		accounts[i] = decredmaterial.TabItem{
			Label: page.theme.Body1(page.info.Wallets[*page.selectedWallet].Accounts[i].Name),
		}
	}

	page.accountsTab.SetTitle(page.theme.Label(values.TextSize18, "Accounts:"))

	page.accountsTab.SetTabs(accounts)
	for range page.accountsTab.ChangeEvent(gtx) {
		*page.selectedAccount = page.accountsTab.Selected
	}

	page.LayoutWithWallets(gtx, func() {
		page.accountsTab.Layout(gtx, body)
	})
}

func toMax(gtx *layout.Context) {
	gtx.Constraints.Width.Min = gtx.Constraints.Width.Max
	gtx.Constraints.Height.Min = gtx.Constraints.Height.Max
}

func mustIcon(ic *decredmaterial.Icon, err error) *decredmaterial.Icon {
	if err != nil {
		panic(err)
	}
	return ic
}