using System;
using Avalonia;
using Avalonia.Automation;
using Avalonia.Controls;
using Avalonia.Controls.Primitives;
using Avalonia.Layout;
using Avalonia.Media;

namespace ECO.UIQualification;

public sealed class MainWindow : Window
{
    private readonly Button _warranty;
    private readonly Button _timeline;
    private readonly TextBox _coreStatus;
    private readonly TextBox _matterTitle;
    private readonly TextBox _coreRevision;
    private readonly TextBox _evidenceSummary;
    private CoreProcessClient? _core;

    public MainWindow()
    {
        Title = "ECO bridge qualification — Avalonia";
        Width = 1000;
        Height = 760;
        MinWidth = 760;
        MinHeight = 560;

        var title = new TextBlock
        {
            Text = "Synthetic Matter Workspace",
            FontSize = 24,
            FontWeight = FontWeight.SemiBold,
            Margin = new Thickness(0, 0, 0, 12)
        };

        _coreStatus = ReadOnlyField("Core status", "Core bridge not connected.");
        _matterTitle = ReadOnlyField("Matter title", "Not loaded");
        _coreRevision = ReadOnlyField("Core revision", "Not loaded");
        _evidenceSummary = ReadOnlyField("Evidence summary", "Not loaded");

        var search = new TextBox
        {
            PlaceholderText = "Search this Matter",
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
            Height = 140,
            HorizontalAlignment = HorizontalAlignment.Stretch
        };
        AutomationProperties.SetName(transcript, "AI conversation transcript");

        search.TextChanged += (_, _) => ApplyFilter(search.Text ?? string.Empty);

        var stack = new StackPanel
        {
            Spacing = 8,
            Margin = new Thickness(24),
            Children =
            {
                title,
                _coreStatus,
                _matterTitle,
                _coreRevision,
                _evidenceSummary,
                search,
                actions,
                transcriptLabel,
                transcript
            }
        };

        Content = new ScrollViewer
        {
            Content = stack,
            HorizontalScrollBarVisibility = ScrollBarVisibility.Disabled,
            VerticalScrollBarVisibility = ScrollBarVisibility.Auto
        };

        Opened += async (_, _) => await ConnectCoreAsync();
        Closing += (_, _) =>
        {
            try { _core?.Dispose(); }
            catch { }
            _core = null;
        };
    }

    private static TextBox ReadOnlyField(string accessibleName, string value)
    {
        var box = new TextBox
        {
            Text = value,
            IsReadOnly = true,
            HorizontalAlignment = HorizontalAlignment.Stretch
        };
        AutomationProperties.SetName(box, accessibleName);
        return box;
    }

    private async System.Threading.Tasks.Task ConnectCoreAsync()
    {
        var exe = Environment.GetEnvironmentVariable("ECO_BRIDGE_CORE_EXE") ?? string.Empty;
        var sha = Environment.GetEnvironmentVariable("ECO_BRIDGE_CORE_SHA256") ?? string.Empty;
        var pidFile = Environment.GetEnvironmentVariable("ECO_BRIDGE_PID_FILE");
        if (string.IsNullOrWhiteSpace(exe))
        {
            _coreStatus.Text = "Core bridge not configured.";
            return;
        }

        try
        {
            _core = CoreProcessClient.Start(exe, sha, pidFile);
            var ping = await _core.PingAsync();
            if (!string.Equals(ping.Status, "succeeded", StringComparison.Ordinal))
                throw new InvalidOperationException("Core ping did not succeed.");
            _coreStatus.Text = ping.UserMessage ?? "ECO core ready.";

            var projected = await _core.ProjectMatterAsync("MAT-SYNTHETIC");
            if (!string.Equals(projected.Status, "succeeded", StringComparison.Ordinal) || projected.Projection is null)
                throw new InvalidOperationException("Matter projection did not succeed.");

            _matterTitle.Text = projected.Projection.Identity.Title;
            _coreRevision.Text = projected.Projection.Revision;
            _evidenceSummary.Text = $"Records {projected.Projection.Evidence.Records}; readable {projected.Projection.Evidence.Readable}; unresolved {projected.Projection.Evidence.Unresolved}";
        }
        catch (Exception ex)
        {
            _coreStatus.Text = "Bridge failed: " + ex.Message;
            try { _core?.Dispose(); } catch { }
            _core = null;
        }
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
