package views

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"
	"log"

	"golang.org/x/term"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"42tui/tui/service"
	"42tui/languages"
)

// profilePicture generates a deterministic 6x6 pixel avatar based on the user's ID hash.
func profilePicture(userID int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d", userID)))
	hash := h[:]

	palette := []string{
		"#5F5FAF", "#FF5FAF", "#7CFF6B", "#FFD166",
		"#06D6A0", "#EF476F", "#A78BFA",
	}

	const width = 6
	const height = 6

	var b strings.Builder

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			idx := int(hash[(i*width+j)%len(hash)]) % len(palette)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[idx]))
			b.WriteString(style.Render("██"))
		}
		if i != height-1 {
			b.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(b.String())
}

// profile layouts the top profile info snippet (Name, Username, Email, Location).
func profile(p service.ProfileData, maxSize int) string {
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5FAF")).Bold(true)
	usernameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff92d5"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6f6f6f"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D0D0D0"))

	fullName := fmt.Sprintf("%s %s", p.FirstName, p.LastName)
	username := "@" + p.Login

	location := p.Location
	if location == "" {
		location = "Not logged."
	}

	header := lipgloss.NewStyle().
		Width(maxSize).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				nameStyle.Render(fullName),
				lipgloss.NewStyle().Width(maxSize-len(fullName)-len(username)).Render(""),
				usernameStyle.Render(username),
			),
		)

	info := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		labelStyle.Render("Email: ")+valueStyle.Render(p.Email),
		labelStyle.Render("Location: ")+valueStyle.Render(location),
		"",
		"",
		fmt.Sprintf(
			"%s %s\t%s %s",
			labelStyle.Render("Ev.P:"),
			valueStyle.Render(fmt.Sprintf("%d",p.EvaluationPoints)),
			labelStyle.Render("Wallet:"),
			valueStyle.Render(fmt.Sprintf("%d",p.Wallet)),
		),
	)

	avatar := profilePicture(p.ID)

	return lipgloss.JoinHorizontal(lipgloss.Top, avatar, " ", info)
}

// Projects renders the vertical table component of submitted academic projects.
func Projects(projects []service.Project) string {
	const (
		nameWidth  = 15
		scoreWidth = 5
	)

	var sb strings.Builder

	for i, p := range projects {
		status := "✘"
		statusColor := lipgloss.Color("#ff4848")
		scoreColor := lipgloss.Color("#ff4848")
		nameColor := lipgloss.Color("#ffffff")

		if p.IsValidated {
			status = "✓"
			statusColor = lipgloss.Color("#2ecc71")
			scoreColor = lipgloss.Color("#9cff67")
		}

		// Zebra striping color rotation for project list
		switch i % 4 {
		case 0:
			nameColor = lipgloss.Color("#F7CAD0")
		case 1:
			nameColor = lipgloss.Color("#DEA6AF")
		case 2:
			nameColor = lipgloss.Color("#B0A8B9")
		case 3:
			nameColor = lipgloss.Color("#BBCCB4")
		}

		nameStyle := lipgloss.NewStyle().Foreground(nameColor)
		scoreStyle := lipgloss.NewStyle().Foreground(scoreColor)
		statusStyle := lipgloss.NewStyle().Foreground(statusColor)

		name := []rune(p.ProjectName)
		if len(name) > nameWidth {
			if nameWidth > 3 {
				name = append(name[:nameWidth-3], []rune("...")...)
			} else {
				name = name[:nameWidth]
			}
		}

		namePart := nameStyle.Width(nameWidth).Render(string(name))
		scorePart := scoreStyle.Width(scoreWidth).Render(fmt.Sprintf("%d", int(p.FinalMark)))
		statusPart := statusStyle.Render(status)

		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, namePart, scorePart, statusPart))
		if i < len(projects)-1 {
			sb.WriteByte('\n')
		}
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5FAF")).
		Bold(true).
		Render("Submitted Projects ⛶")

	return lipgloss.JoinVertical(lipgloss.Center, header, sb.String())
}

// pace tracks curriculum velocity and milestones deadlines.
func pace(p service.ProfileData) string {

	// daysBetween calculates day difference between two strings ("YYYY-MM-DD" or "now").
	daysBetween := func(startStr, endStr string) int {
		const layout = "2006-01-02"
		var startDate time.Time
		var err error

		if startStr == "now" {
			now := time.Now().UTC()
			startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			// Trim full ISO8601 timestamps to basic date length to avoid parsing errors
			if len(startStr) > 10 {
				startStr = startStr[:10]
			}
			startDate, err = time.Parse(layout, startStr)
			if err != nil {
				return 0
			}
		}

		if len(endStr) > 10 {
			endStr = endStr[:10]
		}
		endDate, err := time.Parse(layout, endStr)
		if err != nil {
			return 0
		}

		return int(endDate.Sub(startDate).Hours() / 24)
	}

	currentPace := p.Pace.Pace
	cursusBeginDate := p.Pace.CursusBeginDate
	currentMilestone := p.Pace.Milestone
	currentDeadline := p.Pace.Deadline

	currentMilestoneStartingDate := cursusBeginDate

	if currentMilestone > 0 {
		for _, m := range p.Pace.Milestones {
			if m.Level == currentMilestone-1 && m.ValidatedAt != "" {
				currentMilestoneStartingDate = m.ValidatedAt[:10]
				break
			}
		}
	}

	currentPaceTotalDays := daysBetween(currentMilestoneStartingDate, currentDeadline)
	currentPaceLeftDays := daysBetween("now", currentDeadline)

	// Map milestones by their functional levels to protect against unstructured or unordered API payloads
	milestoneMap := make(map[int]service.Milestone)
	for _, m := range p.Pace.Milestones {
		milestoneMap[m.Level] = m
	}

	if m6, exists := milestoneMap[6]; exists && m6.ValidatedAt != "" {
		currentMilestone = 7 // Status code indicating curriculum completion
	}

	paces := []int{8, 12, 15, 18, 22, 24}

	nextPace := func(current int) int {
		for i, v := range paces {
			if v == current {
				if i+1 < len(paces) {
					return paces[i+1]
				}
				return -1
			}
		}
		return 0
	}

	buildHeader := func() string {

		var nextPaceNumber int
		if p.Pace.Deadline != "" {
			nextPaceNumber = nextPace(currentPace)
		} else {
			return "Cursus completed"
		}

		next := fmt.Sprintf("Pace %d", nextPaceNumber)
		if currentPace == 0 {
			currentPace = 24
			next = "Blackhole"	
		}

		daysLeft := currentPaceLeftDays
		totalDays := currentPaceTotalDays
		daysPassed := totalDays - daysLeft

		if totalDays <= 0 {
			totalDays = 1
		}

		barWidth := 42
		ratio := float64(daysPassed) / float64(totalDays)

		if ratio < 0 {
			ratio = 0
		} else if ratio > 1 {
			ratio = 1
		}

		progressColor := lipgloss.Color("#ff00bf")
		switch {
		case ratio < 0.60: // Green (Safe progress zone)
			progressColor = lipgloss.Color("#7CFF6B")
		case ratio < 0.85: // Orange (Warning zone)
			progressColor = lipgloss.Color("#ffbb4e")
		default: // Red (Critical deadline risk)
			progressColor = lipgloss.Color("#e74c3c")
		}

		filledBlocks := int(ratio * float64(barWidth))
		emptyBlocks := barWidth - filledBlocks

		filledStr := strings.Repeat("█", filledBlocks)
		emptyStr := strings.Repeat("░", emptyBlocks)

		progressStyle := lipgloss.NewStyle().Foreground(progressColor).Render(filledStr)
		bgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7f8c8d")).Render(emptyStr)
		daysLeftStyle := lipgloss.NewStyle().
			Foreground(progressColor).
			Bold(true)

		content := lipgloss.JoinVertical(
			lipgloss.Center,
			fmt.Sprintf(
				"Pace %d [%s%s] %s",
				currentPace,
				progressStyle,
				bgStyle,
				next,
			),
			daysLeftStyle.Render(
				fmt.Sprintf("╰ %d days left ╯", daysLeft),
			),
		)

		return content
	}

	passedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0F172A")). // texto oscuro
		Background(lipgloss.Color("#7CFF6B")). // verde neon suave
		Padding(0, 1)

	currentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#5F5FAF")). // violeta principal
		Bold(true).
		Padding(0, 1)

	pendingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA")). // lila tenue
		Background(lipgloss.Color("#2A2438")). // fondo oscuro elegante
		Padding(0, 1)

	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#bdc3c7")).
		Padding(0, 2)

	durationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#95a5a6")).
		Padding(0, 2)

	var milestones strings.Builder
	var validatedDates strings.Builder
	var durations strings.Builder

	for i := 0; i <= 6; i++ {
		label := fmt.Sprintf("Milestone %d", i)
		var dateLabel string
		var durationLabel string

		if m, exists := milestoneMap[i]; exists {

			if i == currentMilestone {
				dateLabel = "\t  Current"
			} else {
				dateLabel = m.ValidatedAt
				if len(dateLabel) > 10 {
					dateLabel = dateLabel[:10]
				}
				if dateLabel == "" {
					dateLabel = "\t\t---"
				}
			}

			var daysTaken int

			if i == 0 {
				daysTaken = daysBetween(cursusBeginDate, m.ValidatedAt)

			} else {
				if prevM, prevExists := milestoneMap[i-1]; prevExists && prevM.ValidatedAt != "" {
					daysTaken = daysBetween(prevM.ValidatedAt, m.ValidatedAt)
				}
			}

			durationLabel = fmt.Sprintf("Completed in %d days", daysTaken)

		} else {
			dateLabel = "###"
			durationLabel = "###"
		}

		var line string
		var dateLine string
		var durationLine string

		switch {
		case i < currentMilestone:
			line = passedStyle.Render("✔ " + label)
			dateLine = dateStyle.Render("🎉" + " " + dateLabel)
			durationLine = durationStyle.Render(durationLabel)

		case i == currentMilestone:
			line = currentStyle.Render(" ●" + " " + label + " ◀")
			dateLine = dateLabel

			durationLine = fmt.Sprintf("\t%d days passed",
				currentPaceTotalDays - currentPaceLeftDays,
			)
		

		default:
			line = pendingStyle.Render("○ " + label)
			dateLine = dateLabel
			durationLine = "\t\t\t---"
		}

		if i > 0 {
			milestones.WriteByte('\n')
			validatedDates.WriteByte('\n')
			durations.WriteByte('\n')
		}

		milestones.WriteString(line)
		validatedDates.WriteString(dateLine)
		durations.WriteString(durationLine)
	}

	columns := lipgloss.JoinHorizontal(lipgloss.Top,
		milestones.String(),
		validatedDates.String(),
		durations.String(),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		buildHeader(),
		"",
		columns,
	)
}

// Profile aggregates and outputs the root layout structure for the user terminal dashboard viewport.
func Profile(p service.ProfileData, projectsVp viewport.Model) string {

	lan, err := languages.New("es")
	if err != nil {
		log.Printf("Error loading language: %v", err)
		fmt.Printf("Error loading language: %v\n", err)
		os.Exit(1)
	}

	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

	if p.ID == 0 {
		return lipgloss.NewStyle().
			Width(w).
			AlignHorizontal(lipgloss.Center).
			Render("loading")
	}

	if p.ID == -1 {
		return lipgloss.NewStyle().
			Width(w).
			AlignHorizontal(lipgloss.Center).
			Render(lan.Get("warning.freeze"))
	}

	profileInformation := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderRight(false).
		BorderForeground(lipgloss.Color("#FF5FAF")).
		Padding(0, 1).
		Render(profile(p, 40))

	projectsList := lipgloss.NewStyle().Padding(0, 1).Render(projectsVp.View())
	cursusStatus := lipgloss.NewStyle().Padding(2, 1).Render(pace(p))

	profileSection := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinVertical(
			lipgloss.Top,
			profileInformation,
			cursusStatus,
		),
		" ",
		projectsList,
	)

	return lipgloss.NewStyle().
		Width(w).
		AlignHorizontal(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				profileSection,
				"New sections coming soon...",
			),
		)
}