package health

import "fmt"

// maxListedOrphans caps how many names are printed per group in the
// human report. If a token loses access, thousands of mirrors can be
// orphaned at once; the full list is always available via --json.
const maxListedOrphans = 10

// printOrphanNames prints a count and up to maxListedOrphans names,
// noting how many were omitted.
func printOrphanNames(label string, names []string) {

	if len(names) == 0 {
		return
	}

	fmt.Printf("    %s: %d\n", label, len(names))

	shown := names
	if len(shown) > maxListedOrphans {
		shown = names[:maxListedOrphans]
	}

	for _, name := range shown {
		fmt.Printf("      - %s\n", name)
	}

	if omitted := len(names) - len(shown); omitted > 0 {
		fmt.Printf("      ... and %d more (see `gitback health --json`)\n", omitted)
	}
}

// printOrphans prints the "Not in inventory" section, or nothing when
// there are no orphans.
func printOrphans(o OrphanHealth) {

	if o.Mirrors.Count() == 0 && o.Quarantine.Count() == 0 {
		return
	}

	fmt.Println("Not in inventory")
	printOrphanGroup("Live mirrors (no longer updated)", o.Mirrors)
	printOrphanGroup("Quarantined (cannot be recovered automatically)", o.Quarantine)
	fmt.Println()
}

func printOrphanGroup(title string, list OrphanList) {

	if list.Count() == 0 {
		return
	}

	fmt.Printf("  %s\n", title)
	printOrphanNames("Repositories", list.Repositories)
	printOrphanNames("Gists", list.Gists)
}

func PrintReport(report *HealthReport) {

	fmt.Printf("Status: %s\n\n", report.Status)

	fmt.Println("Repositories")
	fmt.Printf("  Healthy: %d\n", report.Repositories.Healthy)
	fmt.Printf("  Failed:  %d\n", report.Repositories.Failed)
	if report.Repositories.Interrupted > 0 {
		fmt.Printf("  Interrupted: %d\n", report.Repositories.Interrupted)
	}
	fmt.Printf("  Total:   %d\n\n", report.Repositories.Total)

	if report.Gists.Total > 0 {
		fmt.Println("Gists")
		fmt.Printf("  Healthy: %d\n", report.Gists.Healthy)
		fmt.Printf("  Failed:  %d\n", report.Gists.Failed)
		if report.Gists.Interrupted > 0 {
			fmt.Printf("  Interrupted: %d\n", report.Gists.Interrupted)
		}
		fmt.Printf("  Total:   %d\n\n", report.Gists.Total)
	}

	if report.Quarantine.Repositories > 0 || report.Quarantine.Gists > 0 {
		fmt.Println("Quarantine")
		fmt.Printf("  Repositories: %d\n", report.Quarantine.Repositories)
		fmt.Printf("  Gists:        %d\n\n", report.Quarantine.Gists)
	}

	// Print orphans
	printOrphans(report.Orphaned)

	fmt.Println("Snapshots")
	fmt.Printf("  Count:  %d\n", report.Snapshots.Count)

	if report.Snapshots.Count > 0 {
		fmt.Printf("  Size:   %s\n", humanSize(int64(report.Snapshots.Size)))
		fmt.Printf("  Latest: %s\n", report.Snapshots.Latest)
	}

	fmt.Println()

	fmt.Println("Storage")

	for _, disk := range report.Disks {

		fmt.Printf("  %s\n", disk.Path)
		fmt.Printf(
			"    Free:  %s (%d%%)\n",
			humanSize(int64(disk.Free)),
			disk.FreePercent,
		)

		fmt.Printf(
			"    Total: %s\n",
			humanSize(int64(disk.Total)),
		)
	}

	if len(report.Warnings) > 0 {

		fmt.Println()

		fmt.Println("Warnings")

		for _, warning := range report.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	if len(report.Recommendations) > 0 {

		fmt.Println()

		fmt.Println("Recommendations")

		for _, recommendation := range report.Recommendations {
			fmt.Printf("  • %s\n", recommendation)
		}
	}
}

func humanSize(b int64) string {

	// Unit names in order.
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	// Work with float64 so we can divide repeatedly.
	size := float64(b)

	// Current unit index.
	// 0 = B
	// 1 = KB
	// 2 = MB
	// ...
	unit := 0

	// Keep dividing by 1024 until the number
	// becomes smaller than 1024.
	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}

	// For bytes, don't show decimal places.
	if unit == 0 {
		return fmt.Sprintf("%d %s", b, units[unit])
	}

	return fmt.Sprintf("%.1f %s", size, units[unit])
}
