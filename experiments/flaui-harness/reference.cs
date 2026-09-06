using System;
using System.Drawing;
using System.Windows.Forms;

internal static class ReferenceApp
{
    [STAThread]
    private static void Main()
    {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);

        var form = new Form
        {
            Text = "ECO FlaUI synthetic reference",
            StartPosition = FormStartPosition.CenterScreen,
            ClientSize = new Size(640, 420)
        };

        var title = new Label
        {
            Text = "Synthetic ECO accessibility reference — no private data",
            AccessibleName = "Synthetic ECO accessibility reference",
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
            Text = "Ready",
            AccessibleName = "Status",
            AutoSize = true,
            Location = new Point(24, 344)
        };

        run.Click += (_, __) => status.Text = "Search complete: " + search.Text;

        form.Controls.AddRange(new Control[] { title, searchLabel, search, run, results, status });
        Application.Run(form);
    }
}
