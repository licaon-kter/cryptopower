package seedbackup

import (
	"math/rand"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"

	"gitlab.com/raedah/cryptopower/app"
	"gitlab.com/raedah/cryptopower/libwallet/wallets/dcr"
	"gitlab.com/raedah/cryptopower/ui/cryptomaterial"
	"gitlab.com/raedah/cryptopower/ui/load"
	"gitlab.com/raedah/cryptopower/ui/modal"
	"gitlab.com/raedah/cryptopower/ui/page/components"
	"gitlab.com/raedah/cryptopower/ui/values"
)

const VerifySeedPageID = "verify_seed"

type shuffledSeedWords struct {
	selectedIndex int
	words         []string
	clickables    []*cryptomaterial.Clickable
}

type VerifySeedPage struct {
	*load.Load
	// GenericPageModal defines methods such as ID() and OnAttachedToNavigator()
	// that helps this Page satisfy the app.Page interface. It also defines
	// helper methods for accessing the PageNavigator that displayed this page
	// and the root WindowNavigator.
	*app.GenericPageModal

	wallet        *dcr.Wallet
	seed          string
	multiSeedList []shuffledSeedWords

	backButton    cryptomaterial.IconButton
	actionButton  cryptomaterial.Button
	listGroupSeed []*layout.List
	list          *widget.List

	redirectCallback Redirectfunc
}

func NewVerifySeedPage(l *load.Load, wallet *dcr.Wallet, seed string, redirect Redirectfunc) *VerifySeedPage {
	pg := &VerifySeedPage{
		Load:             l,
		GenericPageModal: app.NewGenericPageModal(VerifySeedPageID),
		wallet:           wallet,
		seed:             seed,

		actionButton: l.Theme.Button("Verify"),

		redirectCallback: redirect,
	}
	pg.list = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}

	pg.actionButton.Font.Weight = text.Medium

	pg.backButton, _ = components.SubpageHeaderButtons(l)
	pg.backButton.Icon = l.Theme.Icons.ContentClear

	return pg
}

// OnNavigatedTo is called when the page is about to be displayed and
// may be used to initialize page features that are only relevant when
// the page is displayed.
// Part of the load.Page interface.
func (pg *VerifySeedPage) OnNavigatedTo() {
	allSeeds := dcr.PGPWordList()

	listGroupSeed := make([]*layout.List, 0)
	multiSeedList := make([]shuffledSeedWords, 0)
	seedWords := strings.Split(pg.seed, " ")
	rand.Seed(time.Now().UnixNano())
	for _, word := range seedWords {
		listGroupSeed = append(listGroupSeed, &layout.List{Axis: layout.Horizontal})
		index := seedPosition(word, allSeeds)
		shuffledSeed := pg.getMultiSeed(index, dcr.PGPWordList()) // using allSeeds here modifies the slice
		multiSeedList = append(multiSeedList, shuffledSeed)
	}

	pg.multiSeedList = multiSeedList
	pg.listGroupSeed = listGroupSeed
}

func (pg *VerifySeedPage) getMultiSeed(realSeedIndex int, allSeeds []string) shuffledSeedWords {
	shuffledSeed := shuffledSeedWords{
		selectedIndex: -1,
		words:         make([]string, 0),
		clickables:    make([]*cryptomaterial.Clickable, 0),
	}

	clickable := func() *cryptomaterial.Clickable {
		cl := pg.Theme.NewClickable(true)
		cl.Radius = cryptomaterial.Radius(8)
		return cl
	}

	shuffledSeed.words = append(shuffledSeed.words, allSeeds[realSeedIndex])
	shuffledSeed.clickables = append(shuffledSeed.clickables, clickable())
	allSeeds = removeSeed(allSeeds, realSeedIndex)

	for i := 0; i < 3; i++ {
		randomSeed := rand.Intn(len(allSeeds))

		shuffledSeed.words = append(shuffledSeed.words, allSeeds[randomSeed])
		shuffledSeed.clickables = append(shuffledSeed.clickables, clickable())
		allSeeds = removeSeed(allSeeds, randomSeed)
	}

	rand.Shuffle(len(shuffledSeed.words), func(i, j int) {
		shuffledSeed.words[i], shuffledSeed.words[j] = shuffledSeed.words[j], shuffledSeed.words[i]
	})

	return shuffledSeed
}

func seedPosition(seed string, allSeeds []string) int {
	for i := range allSeeds {
		if allSeeds[i] == seed {
			return i
		}
	}
	return -1
}

func removeSeed(allSeeds []string, index int) []string {
	return append(allSeeds[:index], allSeeds[index+1:]...)
}

func (pg *VerifySeedPage) allSeedsSelected() bool {
	for _, multiSeed := range pg.multiSeedList {
		if multiSeed.selectedIndex == -1 {
			return false
		}
	}

	return true
}

func (pg *VerifySeedPage) selectedSeedPhrase() string {
	var wordList []string
	for _, multiSeed := range pg.multiSeedList {
		if multiSeed.selectedIndex != -1 {
			wordList = append(wordList, multiSeed.words[multiSeed.selectedIndex])
		}
	}

	return strings.Join(wordList, " ")
}

func (pg *VerifySeedPage) verifySeed() {
	passwordModal := modal.NewCreatePasswordModal(pg.Load).
		EnableName(false).
		EnableConfirmPassword(false).
		Title("Confirm to verify seed").
		SetPositiveButtonCallback(func(_, password string, m *modal.CreatePasswordModal) bool {
			seed := pg.selectedSeedPhrase()
			_, err := pg.WL.SelectedWallet.Wallet.VerifySeedForWallet(seed, []byte(password))
			if err != nil {
				if err.Error() == dcr.ErrInvalid {
					msg := values.String(values.StrSeedValidationFailed)
					errModal := modal.NewErrorModal(pg.Load, msg, modal.DefaultClickFunc())
					pg.ParentWindow().ShowModal(errModal)
					m.Dismiss()
					return false
				}

				m.SetLoading(false)
				m.SetError(err.Error())
				return false
			}
			m.Dismiss()
			pg.ParentNavigator().Display(NewBackupSuccessPage(pg.Load, pg.redirectCallback))

			return true
		})
	pg.ParentWindow().ShowModal(passwordModal)
}

// HandleUserInteractions is called just before Layout() to determine
// if any user interaction recently occurred on the page and may be
// used to update the page's UI components shortly before they are
// displayed.
// Part of the load.Page interface.
func (pg *VerifySeedPage) HandleUserInteractions() {
	for i, multiSeed := range pg.multiSeedList {
		for j, clickable := range multiSeed.clickables {
			for clickable.Clicked() {
				pg.multiSeedList[i].selectedIndex = j
			}
		}
	}

	for pg.actionButton.Clicked() {
		if pg.allSeedsSelected() {
			pg.verifySeed()
		}
	}
}

// OnNavigatedFrom is called when the page is about to be removed from
// the displayed window. This method should ideally be used to disable
// features that are irrelevant when the page is NOT displayed.
// NOTE: The page may be re-displayed on the app's window, in which case
// OnNavigatedTo() will be called again. This method should not destroy UI
// components unless they'll be recreated in the OnNavigatedTo() method.
// Part of the load.Page interface.
func (pg *VerifySeedPage) OnNavigatedFrom() {}

// Layout draws the page UI components into the provided layout context
// to be eventually drawn on screen.
// Part of the load.Page interface.
func (pg *VerifySeedPage) Layout(gtx C) D {
	if pg.Load.GetCurrentAppWidth() <= gtx.Dp(values.StartMobileView) {
		return pg.layoutMobile(gtx)
	}
	return pg.layoutDesktop(gtx)
}

func (pg *VerifySeedPage) layoutDesktop(gtx layout.Context) layout.Dimensions {
	sp := components.SubPage{
		Load:       pg.Load,
		Title:      "Verify seed word",
		SubTitle:   "Step 2/2",
		BackButton: pg.backButton,
		Back: func() {
			promptToExit(pg.Load, pg.ParentNavigator(), pg.ParentWindow())
		},
		Body: func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					label := pg.Theme.Label(values.TextSize16, "Select the correct words to verify.")
					label.Color = pg.Theme.Color.GrayText1
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Bottom: values.MarginPadding96,
					}.Layout(gtx, func(gtx C) D {
						return pg.Theme.List(pg.list).Layout(gtx, len(pg.multiSeedList), func(gtx C, i int) D {
							return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
								return pg.seedListRow(gtx, i, pg.multiSeedList[i])
							})
						})
					})
				}),
			)
		},
	}

	pg.actionButton.SetEnabled(pg.allSeedsSelected())
	layout := func(gtx C) D {
		return sp.Layout(pg.ParentWindow(), gtx)
	}
	return container(gtx, false, *pg.Theme, layout, "", pg.actionButton)
}

func (pg *VerifySeedPage) layoutMobile(gtx layout.Context) layout.Dimensions {
	sp := components.SubPage{
		Load:       pg.Load,
		Title:      "Verify seed word",
		SubTitle:   "Step 2/2",
		BackButton: pg.backButton,
		Back: func() {
			promptToExit(pg.Load, pg.ParentNavigator(), pg.ParentWindow())
		},
		Body: func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					label := pg.Theme.Label(values.TextSize16, "Select the correct words to verify.")
					label.Color = pg.Theme.Color.GrayText1
					return label.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Bottom: values.MarginPadding96,
					}.Layout(gtx, func(gtx C) D {
						return pg.Theme.List(pg.list).Layout(gtx, len(pg.multiSeedList), func(gtx C, i int) D {
							return pg.seedListRow(gtx, i, pg.multiSeedList[i])
						})
					})
				}),
			)
		},
	}

	pg.actionButton.SetEnabled(pg.allSeedsSelected())
	layout := func(gtx C) D {
		return sp.Layout(pg.ParentWindow(), gtx)
	}
	return container(gtx, true, *pg.Theme, layout, "", pg.actionButton)
}

func (pg *VerifySeedPage) seedListRow(gtx C, index int, multiSeed shuffledSeedWords) D {
	return cryptomaterial.LinearLayout{
		Width:       cryptomaterial.MatchParent,
		Height:      cryptomaterial.WrapContent,
		Orientation: layout.Vertical,
		Background:  pg.Theme.Color.Surface,
		Border:      cryptomaterial.Border{Radius: cryptomaterial.Radius(8)},
		Margin:      layout.Inset{Top: values.MarginPadding4, Bottom: values.MarginPadding4},
		Padding:     layout.Inset{Top: values.MarginPadding16, Right: values.MarginPadding16, Bottom: values.MarginPadding8, Left: values.MarginPadding16},
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			text := "-"
			if multiSeed.selectedIndex != -1 {
				text = multiSeed.words[multiSeed.selectedIndex]
			}
			return seedItem(pg.Theme, gtx, gtx.Constraints.Max.X, index+1, text)
		}),
		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X

			return layout.Inset{Top: values.MarginPadding16}.Layout(gtx, func(gtx C) D {
				widgets := []layout.Widget{
					func(gtx C) D { return pg.seedButton(gtx, 0, multiSeed) },
					func(gtx C) D { return pg.seedButton(gtx, 1, multiSeed) },
					func(gtx C) D { return pg.seedButton(gtx, 2, multiSeed) },
					func(gtx C) D { return pg.seedButton(gtx, 3, multiSeed) },
				}
				return pg.listGroupSeed[index].Layout(gtx, len(widgets), func(gtx C, i int) D {
					return layout.UniformInset(values.MarginPadding0).Layout(gtx, widgets[i])
				})
			})
		}),
	)
}

func (pg *VerifySeedPage) seedButton(gtx C, index int, multiSeed shuffledSeedWords) D {
	borderColor := pg.Theme.Color.Gray2
	textColor := pg.Theme.Color.GrayText2
	if index == multiSeed.selectedIndex {
		borderColor = pg.Theme.Color.Primary
		textColor = pg.Theme.Color.Primary
	}

	return multiSeed.clickables[index].Layout(gtx, func(gtx C) D {

		return cryptomaterial.LinearLayout{
			Width:      gtx.Dp(values.MarginPadding100),
			Height:     gtx.Dp(values.MarginPadding40),
			Background: pg.Theme.Color.Surface,
			Direction:  layout.Center,
			Border:     cryptomaterial.Border{Radius: cryptomaterial.Radius(8), Color: borderColor, Width: values.MarginPadding2},
		}.Layout2(gtx, func(gtx C) D {
			label := pg.Theme.Label(values.TextSize16, multiSeed.words[index])
			label.Color = textColor
			return label.Layout(gtx)
		})
	})
}
