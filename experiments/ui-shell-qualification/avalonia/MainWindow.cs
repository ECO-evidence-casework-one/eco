using System;
using Avalonia;
using Avalonia.Automation;
using Avalonia.Controls;
using Avalonia.Layout;
using Avalonia.Media;

namespace ECO.UIQualification;

public sealed class MainWindow : Window
{
    private readonly Button _warranty;
    private readonly Button _timeline;

    public MainWindow()
    {
        Title = "ECO UI shell qualification — Avalonia";
        Width = 1000;
        Height = 700;
        MinWidth = 760;
        MinHeight = 520;

        var title = new TextBlock
        {
            Text = "Synthetic Matter Workspace",
            FontSize = 24,
            FontWeight = FontWeight.SemiBold,
            Margin = new Thickness(0, 0, 0, 12)
        };

        var search = new TextBox
        {
            Watermark = "Search this Matter",
            HorizontalAlignment = HorizontalAlignment.Stretch,
            Margin = new Thickness(0, 0, 0, 12)
        };
        AutomationProperties.SetName(search, "Search this Matter");

        Button Action(string text)
        {
            var button = new Button { Content = text, Margin = new Thickness(4) };
            AutomationProperties.SetName(button, text);
            return button;
        }

        var addEvidence = Action("Add evidence");
        var reviewSource = Action("Review source details");
        var createTask = Action("Create task");
        var askEco = Action("Ask ECO");
        _warranty = Action("Warranty confirmation");
        _timeline = Action("Build the Matter timeline");

        var actions = new WrapPanel
        {
            Orientation = Orientation.Horizontal,
            HorizontalAlignment = HorizontalAlignment.Stretch,
            Margin = new Thickness(0, 0, 0, 16),
            Children = { addEvidence, reviewSource, createTask, askEco, _warranty, _timeline }
        };

        var transcriptLabel = new TextBlock
        {
            Text = "AI conversation transcript",
            FontWeight = FontWeight.SemiBold,
            Margin = new Thickness(0, 4, 0, 6)
        };

        var transcript = new TextBox
        {
            Text = "Known: warranty confirmation appears in preserved source Email.eml. Source-backed synthetic qualification text only.",
            IsReadOnly = true,
            AcceptsReturn = true,
            TextWrapping = TextWrapping.Wrap,
            Height = 170,
            HorizontalAlignment = HorizontalAlignment.Stretch
        };
        AutomationProperties.SetName(transcript, "AI conversation transcript");

        search.TextChanged += (_, _) => ApplyFilter(search.Text ?? string.Empty);

        var stack = new StackPanel
        {
            Spacing = 8,
            Children = { title, search, actions, transcriptLabel, transcript }
        };

        Content = new ScrollViewer
        {
            Content = stack,
            Padding = new Thickness(24),
            HorizontalScrollBarVisibility = Avalonia.Controls.Primitives.ScrollBarVisibility.Disabled,
            VerticalScrollBarVisibility = Avalonia.Controls.Primitives.ScrollBarVisibility.Auto
        };
    }

    private void ApplyFilter(string value)
    {
        var q = value.Trim();
        if (q.Length == 0)
        {
            _warranty.IsVisible = true;
            _timeline.IsVisible = true;
            return;
        }
        _warranty.IsVisible = "Warranty confirmation".Contains(q, StringComparison.OrdinalIgnoreCase);
        _timeline.IsVisible = "Build the Matter timeline".Contains(q, StringComparison.OrdinalIgnoreCase);
    }
}
