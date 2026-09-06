using System;
using System.Drawing;
using System.Linq;
using System.Windows.Forms;

internal static class ReferenceApp
{
    [STAThread]
    private static void Main(string[] args)
    {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);

        var argumentSelfTest = args.Contains("--arg-self-test", StringComparer.Ordinal);
        var required = new[]
        {
            "--arg-self-test",
            "--test-mode",
            "--test-workspace=C:\\Synthetic Root\\workspace",
            "--test-fixture-root=C:\\Synthetic Root\\fixtures",
            "--test-bridge-token=ECOAccessibilitySyntheticToken20260906R1",
            "--test-bridge-endpoint=C:\\Synthetic Root\\bridge\\endpoint.json"
        };
        var argumentForwardingPassed = !argumentSelfTest || required.All(expected => args.Contains(expected, StringComparer.Ordinal));

        var form = new Form
        {
            Text = argumentSelfTest
                ? (argumentForwardingPassed ? "ECO FlaUI argument forwarding PASS" : "ECO FlaUI argument forwarding FAIL")
                : "ECO FlaUI synthetic reference",
            StartPosition = FormStartPosition.CenterScreen,
            ClientSize = new Size(640, 420)
        };

        var title = new Label
        {
            Text = argumentSelfTest
                ? string.Join("\r\n", args)
                : "Synthetic ECO accessibility reference — no private data",
            AccessibleName = argumentSelfTest ? "Forwarded command line" : "Synthetic ECO accessibility reference",
            AutoSize = true,
            Location = new Point(24, 24)
        };

        var searchLabel = new Label
        {
            Text = "Search whole matter",
            AutoSize = true,
            Location = new Point(24, 72)
        };

        var search = new TextBox
        {
            Text = "warranty confirmation",
            AccessibleName = "Search whole matter",
            Name = "SearchWholeMatter",
            Location = new Point(24, 96),
            Width = 420,
            TabIndex = 0
        };

        var run = new Button
        {
            Text = "Run search",
            AccessibleName = "Run search",
            Name = "RunSearch",
            Location = new Point(460, 94),
            Width = 120,
            TabIndex = 1
        };

        var results = new ListBox
        {
            AccessibleName = "Evidence results",
            Name = "EvidenceResults",
            Location = new Point(24, 150),
            Size = new Size(556, 170),
            TabIndex = 2
        };
        results.Items.AddRange(new object[]
        {
            "Warranty confirmation — synthetic evidence",
            "Alex Rowan — synthetic person",
            "Northvale Devices — synthetic organisation"
        });

        var status = new Label
        {
            Text = argumentSelfTest ? (argumentForwardingPassed ? "Arguments received exactly" : "Arguments missing") : "Ready",
            AccessibleName = "Status",
            AutoSize = true,
            Location = new Point(24, 344)
        };

        run.Click += (_, __) => status.Text = "Search complete: " + search.Text;

        form.Controls.AddRange(new Control[] { title, searchLabel, search, run, results, status });
        Application.Run(form);
    }
}
