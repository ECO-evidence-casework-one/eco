using Avalonia;
using Avalonia.Controls;
using Avalonia.Interactivity;
using Avalonia.Media;
using Avalonia.Threading;
using Eco.Recovery.Models;
using Eco.Recovery.Services;

namespace Eco.Recovery;

public sealed partial class MainWindow : Window
{
    private readonly LocalStore _store = new();
    private readonly RecoveryState _state = new();
    private readonly Dictionary<string, Control> _pages;
    private readonly Dictionary<string, Button> _navigation;
    private bool _fullWorkspace;
    private bool _isInitialized;

    public MainWindow()
    {
        InitializeComponent();

        _pages = new(StringComparer.Ordinal)
        {
            ["CurrentPosition"] = CurrentPositionPage,
            ["Evidence"] = EvidencePage,
            ["Timeline"] = TimelinePage,
            ["Money"] = MoneyPage,
            ["Tasks"] = TasksPage,
            ["Conversations"] = ConversationsPage,
            ["Outputs"] = OutputsPage,
            ["Matters"] = MattersPage,
            ["Settings"] = SettingsPage
        };
        _navigation = new(StringComparer.Ordinal)
        {
            ["CurrentPosition"] = NavCurrent,
            ["Evidence"] = NavEvidence,
            ["Timeline"] = NavTimeline,
            ["Money"] = NavMoney,
            ["Tasks"] = NavTasks,
            ["Conversations"] = NavConversations,
            ["Outputs"] = NavOutputs,
            ["Matters"] = NavMatters,
            ["Settings"] = NavSettings
        };

        SidebarStatus.Text = "Visual recovery baseline • engines not connected";
        SizeChanged += (_, _) => ApplyResponsiveLayout();
        ConversationDraft.GetObservable(TextBox.TextProperty).Subscribe(text =>
        {
            _state.ConversationDraft = text ?? string.Empty;
        });
        _isInitialized = true;
    }

    public void LoadSyntheticMatterForVisuals(string page = "CurrentPosition", double scale = 1.0)
    {
        StartOverlay.IsVisible = false;
        WhatsNewOverlay.IsVisible = false;
        _state.InterfaceScale = scale;
        ApplyScale(scale);
        SetMode(fullWorkspace: false, persist: false);
        SetPage(page, persist: false);
        Toast.IsVisible = false;
    }

    public void OpenStartForVisuals()
    {
        WhatsNewOverlay.IsVisible = false;
        StartOverlay.IsVisible = true;
    }

    public void OpenWhatsNewForVisuals()
    {
        StartOverlay.IsVisible = false;
        WhatsNewOverlay.IsVisible = true;
    }

    private async void OnCreateGuidedMatter(object? sender, RoutedEventArgs e)
    {
        StartOverlay.IsVisible = false;
        SetMode(fullWorkspace: false, persist: false);
        SetPage("CurrentPosition", persist: false);
        await PersistAsync();
        ShowToast("Guided Matter created from synthetic information");
    }

    private async void OnCreateFullMatter(object? sender, RoutedEventArgs e)
    {
        StartOverlay.IsVisible = false;
        SetMode(fullWorkspace: true, persist: false);
        SetPage("CurrentPosition", persist: false);
        await PersistAsync();
        ShowToast("Full workspace opened with synthetic information");
    }

    private async void OnOpenExisting(object? sender, RoutedEventArgs e)
    {
        var loaded = await _store.LoadAsync();
        if (loaded is null)
        {
            ShowToast("No saved recovery Matter exists yet");
            return;
        }

        _state.ViewMode = loaded.ViewMode;
        _state.ActivePage = loaded.ActivePage;
        _state.ActiveConversation = loaded.ActiveConversation;
        _state.ConversationDraft = loaded.ConversationDraft;
        _state.InterfaceScale = loaded.InterfaceScale;
        _state.LowSensory = loaded.LowSensory;
        ConversationDraft.Text = loaded.ConversationDraft;
        LowSensoryCheck.IsChecked = loaded.LowSensory;
        ApplyScale(loaded.InterfaceScale);
        SetMode(string.Equals(loaded.ViewMode, "Full", StringComparison.Ordinal), persist: false);
        SetPage(_pages.ContainsKey(loaded.ActivePage) ? loaded.ActivePage : "CurrentPosition", persist: false);
        StartOverlay.IsVisible = false;
        ShowToast("Saved recovery Matter reopened explicitly");
    }

    private void OnShowStart(object? sender, RoutedEventArgs e)
    {
        StartOverlay.IsVisible = true;
    }

    private async void OnNavigate(object? sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string page })
        {
            StartOverlay.IsVisible = false;
            SetPage(page, persist: false);
            await PersistAsync();
        }
    }

    private void SetPage(string page, bool persist)
    {
        if (!_pages.ContainsKey(page))
        {
            page = "CurrentPosition";
        }

        foreach (var pair in _pages)
        {
            pair.Value.IsVisible = string.Equals(pair.Key, page, StringComparison.Ordinal);
        }

        foreach (var pair in _navigation)
        {
            if (string.Equals(pair.Key, page, StringComparison.Ordinal))
            {
                if (!pair.Value.Classes.Contains("Selected"))
                {
                    pair.Value.Classes.Add("Selected");
                }
            }
            else
            {
                pair.Value.Classes.Remove("Selected");
            }
        }

        _state.ActivePage = page;
        PageTitle.Text = page switch
        {
            "CurrentPosition" => "Current Position",
            "Conversations" => "Conversations",
            "Settings" => "Trust & Settings",
            "Matters" => "All Matters",
            _ => page
        };

        if (persist)
        {
            _ = PersistAsync();
        }
    }

    private async void OnToggleMode(object? sender, RoutedEventArgs e)
    {
        SetMode(!_fullWorkspace, persist: false);
        await PersistAsync();
    }

    private void OnSetGuided(object? sender, RoutedEventArgs e)
    {
        SetMode(fullWorkspace: false, persist: true);
    }

    private void OnSetFull(object? sender, RoutedEventArgs e)
    {
        SetMode(fullWorkspace: true, persist: true);
    }

    private void SetMode(bool fullWorkspace, bool persist)
    {
        _fullWorkspace = fullWorkspace;
        _state.ViewMode = fullWorkspace ? "Full" : "Guided";
        ModeButton.Content = fullWorkspace ? "Guided View" : "Full Workspace";
        ModeButton.SetValue(AutomationProperties.NameProperty, fullWorkspace ? "Switch to Guided View" : "Switch to Full Workspace");
        SetButtonEmphasis(GuidedModeSetting, !fullWorkspace);
        SetButtonEmphasis(FullModeSetting, fullWorkspace);
        ShowToast(fullWorkspace ? "Full Workspace selected" : "Guided View selected");

        if (persist)
        {
            _ = PersistAsync();
        }
    }

    private static void SetButtonEmphasis(Button button, bool primary)
    {
        button.Classes.Remove("Primary");
        button.Classes.Remove("Secondary");
        button.Classes.Add(primary ? "Primary" : "Secondary");
    }

    private void OnWhatsNew(object? sender, RoutedEventArgs e)
    {
        WhatsNewOverlay.IsVisible = true;
    }

    private void OnCloseWhatsNew(object? sender, RoutedEventArgs e)
    {
        WhatsNewOverlay.IsVisible = false;
    }

    private void OnOpenSyntheticSource(object? sender, RoutedEventArgs e)
    {
        SetPage("Evidence", persist: true);
        ShowToast("Exact synthetic source opened in the Evidence viewer");
    }

    private void OnCompareSources(object? sender, RoutedEventArgs e)
    {
        SetPage("Timeline", persist: true);
        ShowToast("Both synthetic source accounts are shown without forcing agreement");
    }

    private void OnRecordUnavailable(object? sender, RoutedEventArgs e)
    {
        ShowToast("Unavailable-evidence note control reached");
    }

    private void OnSyntheticImport(object? sender, RoutedEventArgs e)
    {
        ShowToast("Evidence intake integration is the next functional gate");
    }

    private void OnAddNote(object? sender, RoutedEventArgs e)
    {
        ShowToast("Source note control reached");
    }

    private void OnAddTask(object? sender, RoutedEventArgs e)
    {
        ShowToast("Task creation control reached");
    }

    private void OnNewConversation(object? sender, RoutedEventArgs e)
    {
        ShowToast("Conversation tabs are visible; durable multi-chat wiring is next");
    }

    private void OnSendConversation(object? sender, RoutedEventArgs e)
    {
        if (string.IsNullOrWhiteSpace(ConversationDraft.Text))
        {
            ShowToast("Type a question or drafting request first");
            return;
        }

        AIStatusText.Text = "Offline engine integration pending";
        StopButton.IsEnabled = false;
        ShowToast("No fake answer was generated: offline engine is not connected in this visual gate");
        _ = PersistAsync();
    }

    private void OnStopAI(object? sender, RoutedEventArgs e)
    {
        AIStatusText.Text = "Stopped";
        StopButton.IsEnabled = false;
        ShowToast("Generation stopped");
    }

    private void OnOpenDraft(object? sender, RoutedEventArgs e)
    {
        SetPage("Conversations", persist: true);
        ShowToast("Draft conversation opened");
    }

    private void OnNotImplemented(object? sender, RoutedEventArgs e)
    {
        ShowToast("This synthetic isolation Matter is not opened in the first visual slice");
    }

    private void OnScaleChanged(object? sender, SelectionChangedEventArgs e)
    {
        if (!_isInitialized)
        {
            return;
        }

        if (ScaleSelector.SelectedItem is ComboBoxItem { Tag: string raw } &&
            double.TryParse(raw, System.Globalization.NumberStyles.Float, System.Globalization.CultureInfo.InvariantCulture, out var scale))
        {
            _state.InterfaceScale = scale;
            ApplyScale(scale);
            _ = PersistAsync();
            ShowToast($"Interface scale set to {scale * 100:0}%");
        }
    }

    private void ApplyScale(double scale)
    {
        scale = Math.Clamp(scale, 1.0, 2.25);
        FontSize = 14 * Math.Min(scale, 1.5);
        _state.InterfaceScale = scale;

        var selectedIndex = scale switch
        {
            >= 2.20 => 3,
            >= 1.90 => 2,
            >= 1.40 => 1,
            _ => 0
        };
        if (ScaleSelector.SelectedIndex != selectedIndex)
        {
            ScaleSelector.SelectedIndex = selectedIndex;
        }

        ApplyResponsiveLayout();
    }

    private void OnLowSensoryChanged(object? sender, RoutedEventArgs e)
    {
        _state.LowSensory = LowSensoryCheck.IsChecked == true;
        _ = PersistAsync();
        ShowToast(_state.LowSensory ? "Reduced-motion preference recorded" : "Standard motion preference recorded");
    }

    private void ApplyResponsiveLayout()
    {
        var highScale = _state.InterfaceScale >= 1.9;
        var compact = Bounds.Width < 1050 && !highScale;
        AppRoot.ColumnDefinitions[0].Width = new GridLength(highScale ? 280 : compact ? 196 : 248);
        BrandSubtitle.IsVisible = !compact;
    }

    private async Task PersistAsync()
    {
        try
        {
            await _store.SaveAsync(_state);
        }
        catch (Exception ex)
        {
            SidebarStatus.Text = "Local save failed • earlier information remains available";
            ShowToast($"Local save stopped safely: {ex.GetType().Name}");
        }
    }

    private void ShowToast(string message)
    {
        ToastText.Text = message;
        Toast.IsVisible = true;
        DispatcherTimer.RunOnce(() => Toast.IsVisible = false, TimeSpan.FromSeconds(3));
    }
}
