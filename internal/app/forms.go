package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"lazybeads/internal/beads"
)

func (m *Model) updateForm(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch m.formFocus {
	case 0:
		var cmd tea.Cmd
		m.formTitle, cmd = m.formTitle.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		var cmd tea.Cmd
		m.formDesc, cmd = m.formDesc.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		// Priority selection
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "left", "h":
				if m.formPriority > 0 {
					m.formPriority--
				}
			case "right", "l":
				if m.formPriority < 4 {
					m.formPriority++
				}
			}
		}
	case 3:
		// Type selection
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			types := []string{"task", "bug", "feature", "epic", "chore"}
			idx := 0
			for i, t := range types {
				if t == m.formType {
					idx = i
					break
				}
			}
			switch keyMsg.String() {
			case "left", "h":
				idx = (idx - 1 + len(types)) % len(types)
			case "right", "l":
				idx = (idx + 1) % len(types)
			}
			m.formType = types[idx]
		}
	case 4:
		if m.formParentManual {
			var cmd tea.Cmd
			m.formParent, cmd = m.formParent.Update(msg)
			cmds = append(cmds, cmd)
		} else if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// index 0 = "(none)", 1..N = epics, N+1 = "(enter manually)"
			optionCount := len(m.formParentEpics) + 2
			switch keyMsg.String() {
			case "up", "k":
				if m.formParentIdx > 0 {
					m.formParentIdx--
				}
			case "down", "j":
				if m.formParentIdx < optionCount-1 {
					m.formParentIdx++
				}
			case "enter":
				if m.formParentIdx == optionCount-1 {
					m.formParentManual = true
					m.formParent.SetValue("")
					m.formParent.Focus()
				}
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) resetForm() {
	m.formTitle.SetValue("")
	m.formDesc.SetValue("")
	m.formPriority = 2
	m.formType = "feature"
	m.formParent.SetValue("")
	m.formParentIdx = 0
	m.formParentManual = false
	m.formFocus = 0
	m.updateFormFocus()
}

func (m *Model) updateFormFocus() {
	m.formTitle.Blur()
	m.formDesc.Blur()
	m.formParent.Blur()
	switch m.formFocus {
	case 0:
		m.formTitle.Focus()
	case 1:
		m.formDesc.Focus()
	case 4:
		if m.formParentManual {
			m.formParent.Focus()
		}
	}
}

func (m *Model) submitForm() tea.Cmd {
	title := strings.TrimSpace(m.formTitle.Value())
	if title == "" {
		m.err = fmt.Errorf("title is required")
		return nil
	}

	if m.editing {
		return func() tea.Msg {
			err := m.client.Update(m.editingID, beads.UpdateOptions{
				Title:    title,
				Priority: &m.formPriority,
			})
			return taskUpdatedMsg{err: err}
		}
	}

	var parent string
	if m.formParentManual {
		parent = strings.TrimSpace(m.formParent.Value())
	} else if m.formParentIdx > 0 && m.formParentIdx <= len(m.formParentEpics) {
		parent = m.formParentEpics[m.formParentIdx-1].ID
	}
	return func() tea.Msg {
		task, err := m.client.Create(beads.CreateOptions{
			Title:       title,
			Description: m.formDesc.Value(),
			Type:        m.formType,
			Priority:    m.formPriority,
			Parent:      parent,
		})
		return taskCreatedMsg{task: task, err: err}
	}
}
