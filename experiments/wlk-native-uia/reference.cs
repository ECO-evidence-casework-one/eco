using System;
using System.Drawing;
using System.Windows.Forms;

public sealed class EcoReferenceForm : Form
{
    public EcoReferenceForm()
    {
        Text = "ECO UIA runner reference";
        Width = 700;
        Height = 460;
        StartPosition = FormStartPosition.CenterScreen;

        var label = new Label {
            Text = "Synthetic accessibility runner reference",
            AutoSize = true,
            Left = 20,
            Top = 20,
            AccessibleName = "Synthetic accessibility runner reference"
        };
        var search = new TextBox {
            Text = "warranty confirmation",
            Left = 20,
            Top = 60,
            Width = 430,
            AccessibleName = "Reference search",
            AccessibleDescription = "Known-good Windows Forms edit reference"
        };
        var run = new Button {
            Text = "Reference run",
            Left = 470,
            Top = 58,
            Width = 150,
            AccessibleName = "Reference run",
            AccessibleDescription = "Known-good Windows Forms button reference"
        };
        var results = new ListBox {
            Left = 20,
            Top = 110,
            Width = 600,
            Height = 220,
            AccessibleName = "Reference results",
            AccessibleDescription = "Known-good Windows Forms list reference"
        };
        results.Items.Add("Warranty confirmation — synthetic evidence");
        results.Items.Add("Alex Rowan — synthetic person");
        results.Items.Add("Northvale Devices — synthetic organisation");
        var status = new Label {
            Text = "Reference ready",
            AutoSize = true,
            Left = 20,
            Top = 350
        };
        run.Click += delegate { status.Text = "Reference complete: " + search.Text; };
        Controls.Add(label);
        Controls.Add(search);
        Controls.Add(run);
        Controls.Add(results);
        Controls.Add(status);
    }

    [STAThread]
    public static void Main()
    {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);
        Application.Run(new EcoReferenceForm());
    }
}
