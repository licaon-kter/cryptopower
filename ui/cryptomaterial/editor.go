// SPDX-License-Identifier: Unlicense OR MIT

package cryptomaterial

import (
	"image"
	"image/color"
	"io"
	"strings"

	"gioui.org/gesture"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github.com/crypto-power/cryptopower/ui/values"
)

type RestoreEditor struct {
	t          *Theme
	Edit       *Editor
	TitleLabel Label
	LineColor  color.NRGBA
	height     int
}

type Editor struct {
	t *Theme
	material.EditorStyle

	TitleLabel Label
	errorLabel Label
	LineColor  color.NRGBA

	// IsRequired if true, displays a required field text at the bottom of the
	// editor.
	IsRequired bool
	// IsTitleLabel if true makes the title label visible.
	IsTitleLabel bool

	HasCustomButton bool
	CustomButton    Button

	// Set ExtraText to show a custom text in the editor.
	ExtraText string
	// Bordered if true makes the adds a border around the editor.
	Bordered bool
	// isPassword if true, displays the show and hide button.
	isPassword bool
	// If showEditorIcon is true, displays the editor widget Icon of choice
	showEditorIcon     bool
	alignEditorIconEnd bool

	// isEditorButtonClickable passes a clickable icon button if true and regular icon if false
	isEditorButtonClickable bool

	requiredErrorText string

	editorIcon       *Icon
	editorIconButton IconButton
	showHidePassword IconButton

	EditorIconButtonEvent func()

	m2 unit.Dp
	m5 unit.Dp

	click gesture.Click

	copy, paste   Button
	isDisableMenu bool
	isShowMenu    bool
	focused       bool

	// add space for error lable if it is true
	isSpaceError bool

	isFirstFocus bool

	submitted       bool
	changed         bool
	selected        bool
	showHintOnFocus bool
}

func (t *Theme) EditorPassword(editor *widget.Editor, hint string) Editor {
	editor.Mask = '*'
	e := t.Editor(editor, hint)
	e.isPassword = true
	e.showEditorIcon = false
	return e
}

func (t *Theme) RestoreEditor(editor *widget.Editor, hint string, title string) *RestoreEditor {
	e := t.Editor(editor, hint)
	e.Bordered = false
	e.SelectionColor = color.NRGBA{}
	return &RestoreEditor{
		t:          t,
		Edit:       &e,
		TitleLabel: t.Body2(title),
		LineColor:  t.Color.Gray2,
		height:     31,
	}
}

// IconEditor creates an editor widget with icon of choice
func (t *Theme) IconEditor(editor *widget.Editor, hint string, icon *widget.Icon, clickableIcon bool) Editor {
	e := t.Editor(editor, hint)
	e.showEditorIcon = true
	e.isEditorButtonClickable = clickableIcon
	e.editorIcon = NewIcon(icon)
	e.editorIcon.Color = t.Color.Gray1
	e.editorIconButton.IconButtonStyle.Icon = icon
	return e
}

func (t *Theme) SearchEditor(editor *widget.Editor, hint string, icon *widget.Icon) Editor {
	e := t.Editor(editor, hint)
	e.showEditorIcon = true
	e.editorIcon = NewIcon(icon)
	e.editorIcon.Color = t.Color.Gray1
	e.editorIconButton.IconButtonStyle.Icon = icon
	e.alignEditorIconEnd = false
	e.IsTitleLabel = false
	return e
}

func (t *Theme) Editor(editor *widget.Editor, hint string) Editor {
	errorLabel := t.Caption("")
	errorLabel.Color = t.Color.Danger

	m := material.Editor(t.Base, editor, hint)
	m.TextSize = t.TextSize
	m.Color = t.Color.Text
	m.HintColor = t.Color.GrayText3

	m0 := unit.Dp(0)

	newEditor := Editor{
		t:            t,
		EditorStyle:  m,
		TitleLabel:   t.Body2(""),
		IsTitleLabel: true,
		Bordered:     true,

		alignEditorIconEnd: true,

		errorLabel:        errorLabel,
		requiredErrorText: "Field is required",

		m2: unit.Dp(2),
		m5: unit.Dp(5),

		editorIconButton: IconButton{
			IconButtonStyle{
				Size:   values.MarginPadding24,
				Inset:  layout.UniformInset(m0),
				Button: new(widget.Clickable),
			},
			t.Styles.IconButtonColorStyle, // automatically changes on theme change, to use fixed colors, pass a &values.ColorStyle{} instead.
		},
		showHidePassword: IconButton{
			IconButtonStyle{
				Size:   values.MarginPadding24,
				Inset:  layout.UniformInset(m0),
				Button: new(widget.Clickable),
			},
			t.Styles.IconButtonColorStyle,
		},
		CustomButton: t.Button(""),
		copy:         t.Button(values.String(values.StrCopy)),
		paste:        t.Button(values.String(values.StrPaste)),
	}
	newEditor.copy.TextSize = values.TextSize10
	newEditor.copy.Background = color.NRGBA{}
	newEditor.copy.HighlightColor = t.Color.SurfaceHighlight
	newEditor.copy.Color = t.Color.Primary
	newEditor.copy.Inset = layout.UniformInset(values.MarginPadding5)

	newEditor.paste.TextSize = values.TextSize10
	newEditor.paste.Background = color.NRGBA{}
	newEditor.paste.HighlightColor = t.Color.SurfaceHighlight
	newEditor.paste.Color = t.Color.Primary
	newEditor.paste.Inset = layout.UniformInset(values.MarginPadding5)
	t.allEditors = append(t.allEditors, &newEditor)

	return newEditor
}

func (e *Editor) Pressed(gtx C) bool {
	return e.click.Pressed() || e.copy.Clicked(gtx) || e.paste.Clicked(gtx)
}

func (e *Editor) FirstPressed(gtx C) bool {
	return !gtx.Source.Focused(e.Editor) && e.click.Pressed()

}

func (e *Editor) IsFocused() bool {
	return e.focused
}

func (e *Editor) SetFocus() {
	e.isFirstFocus = true
}

func (e *Editor) UpdateFocus(focus bool) {
	e.isFirstFocus = focus
}

func (e *Editor) Changed() bool {
	changed := e.changed
	e.changed = false
	return changed
}

func (e *Editor) Submitted() bool {
	submitted := e.submitted
	e.submitted = false
	return submitted
}

func (e *Editor) Selected() bool {
	selected := e.selected
	e.selected = false
	return selected
}

func (e *Editor) AlwaysShowHint() {
	e.showHintOnFocus = true
}

func (e *Editor) Layout(gtx C) D {
	if e.isFirstFocus {
		e.isFirstFocus = false
		gtx.Execute(key.FocusCmd{Tag: e.Editor})
	}
	e.handleEvents(gtx)
	e.update(gtx)
	return e.layout(gtx)
}

func (e *Editor) update(gtx C) {
	for {
		ev, ok := e.click.Update(gtx.Source)
		if !ok {
			break
		}
		if e.click.Pressed() {
			if ev.NumClicks > 1 || (e.focused && !e.isShowMenu) {
				e.isShowMenu = true
			} else {
				e.isShowMenu = false
			}
		}
	}

	e.focused = gtx.Source.Focused(e.Editor)

	for {
		ev, ok := e.Editor.Update(gtx)
		if !ok {
			break
		}

		switch ev.(type) {
		case widget.ChangeEvent:
			e.changed = true
		case widget.SubmitEvent:
			e.submitted = true
		case widget.SelectEvent:
			e.selected = true
		}
	}
}

func (e *Editor) layout(gtx C) D {
	e.LineColor, e.TitleLabel.Color = e.t.Color.Gray2, e.t.Color.GrayText3
	if e.Editor.Len() > 0 && len(e.Hint) > 0 {
		e.TitleLabel.Text = e.Hint
	} else if e.Hint == "" {
		e.Hint = e.TitleLabel.Text
		e.TitleLabel.Text = ""
	}

	focused := gtx.Source.Focused(e.Editor)
	if focused {
		// Only non-read only editors should indicate an active state on focus.
		if !e.Editor.ReadOnly {
			e.LineColor = e.t.Color.Primary
		}
		e.TitleLabel.Color = e.t.Color.Primary
		e.TitleLabel.Text = e.Hint
		if !e.showHintOnFocus {
			e.Hint = ""
		}
	}

	if e.IsRequired && !focused && e.Editor.Len() == 0 {
		e.errorLabel.Text = e.requiredErrorText
		e.LineColor = e.t.Color.Danger
	}

	if e.errorLabel.Text != "" {
		e.LineColor, e.TitleLabel.Color = e.t.Color.Danger, e.t.Color.Danger
	}

	overLay := func(_ C) D { return D{} }
	if e.Editor.ReadOnly {
		overLay = func(gtx C) D {
			gtxCopy := gtx
			gtxCopy.Constraints.Max.Y = gtx.Dp(values.MarginPadding46)
			return DisableLayout(nil, gtxCopy, nil, nil, 20, e.t.Color.Gray3, nil)
		}
		gtx = gtx.Disabled()
	}

	dims := layout.UniformInset(e.m2).Layout(gtx, func(gtx C) D {
		return Card{Color: e.t.Color.Surface, Radius: Radius(8)}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Stack{}.Layout(gtx,
						layout.Stacked(func(gtx C) D {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(e.editorLayout),
								layout.Rigid(func(gtx C) D {
									if e.errorLabel.Text != "" {
										inset := layout.Inset{
											Top:  e.m2,
											Left: e.m5,
										}
										return inset.Layout(gtx, e.errorLabel.Layout)
									}
									if e.isSpaceError {
										return layout.Spacer{Height: values.MarginPadding18}.Layout(gtx)
									}
									return D{}
								}),
							)
						}),
						layout.Stacked(func(gtx C) D {
							if e.IsTitleLabel {
								return layout.Inset{
									Top:  values.MarginPaddingMinus10,
									Left: values.MarginPadding10,
								}.Layout(gtx, func(gtx C) D {
									return Card{Color: e.t.Color.Surface}.Layout(gtx, e.TitleLabel.Layout)
								})
							}
							return D{}
						}),
						layout.Expanded(func(gtx C) D {
							defer pointer.PassOp{}.Push(gtx.Ops).Pop()
							defer clip.Rect(image.Rectangle{
								Max: gtx.Constraints.Min,
							}).Push(gtx.Ops).Pop()
							e.click.Add(gtx.Ops)
							return D{}
						}),
						layout.Stacked(overLay),
					)
				}),
			)
		})
	})
	if !e.isDisableMenu {
		e.editorMenusLayout(gtx, dims.Size.Y)
	}
	return dims
}

func (e *Editor) editorLayout(gtx C) D {
	if e.Bordered {
		border := widget.Border{Color: e.LineColor, CornerRadius: unit.Dp(8), Width: unit.Dp(2)}
		return border.Layout(gtx, func(gtx C) D {
			inset := layout.Inset{
				Top:    values.MarginPadding7,
				Bottom: values.MarginPadding7,
				Left:   values.MarginPadding12,
				Right:  values.MarginPadding12,
			}
			return inset.Layout(gtx, e.editor)
		})
	}

	inset := layout.Inset{
		Top:    values.MarginPadding3,
		Bottom: values.MarginPadding3,
		Left:   values.MarginPadding12,
		Right:  values.MarginPadding12,
	}
	return inset.Layout(gtx, e.editor)
}

func (e *Editor) editorMenusLayout(gtx C, editorHeight int) {
	e.isShowMenu = e.isShowMenu && (gtx.Source.Focused(e.Editor) || e.copy.Hovered() || e.paste.Hovered())
	if e.isShowMenu {
		flexChilds := make([]layout.FlexChild, 0)
		if len(e.Editor.SelectedText()) > 0 {
			flexChilds = append(flexChilds, layout.Rigid(e.copy.Layout))
			flexChilds = append(flexChilds, layout.Rigid(e.t.Line(20, 1).Layout))
		}
		flexChilds = append(flexChilds, layout.Rigid(e.paste.Layout))
		gtxCopy := gtx
		macro := op.Record(gtxCopy.Ops)
		LinearLayout{
			Width:      WrapContent,
			Height:     WrapContent,
			Background: e.t.Color.Surface,
			Margin:     layout.Inset{Top: gtxCopy.Metric.PxToDp(-(editorHeight - 10))},
			Padding:    layout.UniformInset(values.MarginPadding5),
			Alignment:  layout.Middle,
			Border:     Border{Radius: Radius(8), Color: e.t.Color.Gray2, Width: unit.Dp(0.5)},
		}.Layout(gtxCopy, flexChilds...)
		op.Defer(gtxCopy.Ops, macro.Stop())
	}
}

func (e *Editor) layoutIconEditor(gtx C) D {
	inset := layout.Inset{
		Top: e.m2,
	}

	if e.alignEditorIconEnd {
		inset.Left = e.m5
	} else {
		inset.Right = e.m5
	}

	return inset.Layout(gtx, func(gtx C) D {
		if e.isEditorButtonClickable {
			return e.editorIconButton.Layout(gtx)
		}
		return e.editorIcon.Layout(gtx, unit.Dp(25))
	})
}

func (e *Editor) editor(gtx C) D {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if e.showEditorIcon && !e.alignEditorIconEnd {
				return e.layoutIconEditor(gtx)
			}
			return D{}
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Inset{Top: e.m5, Bottom: e.m5}.Layout(gtx, e.EditorStyle.Layout)
		}),
		layout.Rigid(func(gtx C) D {
			if e.ExtraText == "" {
				return D{}
			}
			return layout.Inset{Top: values.MarginPadding5}.Layout(gtx, e.t.Label(values.TextSize16, e.ExtraText).Layout)
		}),
		layout.Rigid(func(gtx C) D {
			if e.showEditorIcon && e.alignEditorIconEnd {
				return e.layoutIconEditor(gtx)
			} else if e.isPassword {
				inset := layout.Inset{
					Top:  e.m2,
					Left: e.m5,
				}
				return inset.Layout(gtx, func(gtx C) D {
					icon := MustIcon(widget.NewIcon(icons.ActionVisibilityOff))
					if e.Editor.Mask == '*' {
						icon = MustIcon(widget.NewIcon(icons.ActionVisibility))
					}
					e.showHidePassword.Icon = icon
					return e.showHidePassword.Layout(gtx)
				})
			}
			return D{}
		}),
		layout.Rigid(func(gtx C) D {
			if e.HasCustomButton {
				inset := layout.Inset{
					Top:   e.m5,
					Left:  e.m5,
					Right: e.m5,
				}
				return inset.Layout(gtx, func(gtx C) D {
					e.CustomButton.TextSize = unit.Sp(10)
					return e.CustomButton.Layout(gtx)
				})
			}
			return D{}
		}),
	)
}

func (e *Editor) handleEvents(gtx C) {
	if e.showHidePassword.Button.Clicked(gtx) {
		if e.Editor.Mask == '*' {
			e.Editor.Mask = 0
		} else if e.Editor.Mask == 0 {
			e.Editor.Mask = '*'
		}
	}

	if e.editorIconButton.Button.Clicked(gtx) {
		e.EditorIconButtonEvent()
	}

	if e.copy.Clicked(gtx) {
		gtx.Execute(clipboard.WriteCmd{Data: io.NopCloser(strings.NewReader(e.Editor.SelectedText()))})
		e.isShowMenu = false
	}

	if e.paste.Clicked(gtx) {
		gtx.Execute(clipboard.ReadCmd{Tag: e.Editor})
		e.isShowMenu = false
	}
}

func (re RestoreEditor) Layout(gtx C) D {
	width := int(gtx.Metric.PxPerDp * 2.0)
	height := int(gtx.Metric.PxPerDp * float32(re.height))
	l := re.t.SeparatorVertical(height, width)
	if gtx.Source.Focused(re.Edit.Editor) {
		re.TitleLabel.Color, re.LineColor, l.Color = re.t.Color.Primary, re.t.Color.Primary, re.t.Color.Primary
	} else {
		l.Color = re.t.Color.Gray2
	}
	border := widget.Border{Color: re.LineColor, CornerRadius: values.MarginPadding8, Width: values.MarginPadding2}
	return border.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(values.MarginPadding40)
				return layout.Center.Layout(gtx, re.TitleLabel.Layout)
			}),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: unit.Dp(-3), Right: unit.Dp(5)}.Layout(gtx, l.Layout)
			}),
			layout.Rigid(func(gtx C) D {
				edit := re.Edit.Layout(gtx)
				re.height = edit.Size.Y
				return edit
			}),
		)
	})
}

func (e *Editor) SetRequiredErrorText(txt string) {
	e.requiredErrorText = txt
}

func (e *Editor) SetError(text string) {
	e.errorLabel.Text = text
}

func (e *Editor) HasError() bool {
	return e.errorLabel.Text != ""
}

func (e *Editor) ClearError() {
	e.errorLabel.Text = ""
}

func (e *Editor) IsDirty() bool {
	return e.errorLabel.Text == ""
}

func (e *Editor) AllowSpaceError(allow bool) {
	e.isSpaceError = allow
}
