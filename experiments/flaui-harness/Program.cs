using System.Text.Json;
using FlaUI.Core;
using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using FlaUI.UIA3;

namespace Eco.UiaAcceptance;

internal sealed class AcceptanceProfile
{
    public string? WindowTitleContains { get; set; }
    public List<ControlExpectation> Controls { get; set; } = new();
}

internal sealed class ControlExpectation
{
    public string Name { get; set; } = "";
    public string? AutomationId { get; set; }
    public string ControlType { get; set; } = "Custom";
    public bool MustBeEnabled { get; set; } = true;
    public bool? MustBeKeyboardFocusable { get; set; }
    public List<string> RequiredPatterns { get; set; } = new();
    public string? Exercise { get; set; }
    public string? ExerciseValue { get; set; }
    public int? SelectIndex { get; set; }
}

internal sealed record ControlReceipt(
    string Name,
    string AutomationId,
    string ControlType,
    bool Enabled,
    bool KeyboardFocusable,
    IReadOnlyList<string> SupportedPatterns,
    string? Exercise,
    string Result);

internal sealed record AcceptanceReceipt(
    string Target,
    DateTimeOffset TimestampUtc,
    string WindowTitle,
    string FlaUiCoreVersion,
    string FlaUiUia3Version,
    IReadOnlyList<ControlReceipt> Controls,
    string OverallResult);

internal static class Program
{
    private static int Main(string[] args)
    {
        try
        {
            var options = ParseArgs(args);
            var exe = Require(options, "exe");
            var profilePath = Require(options, "profile");
            var receiptPath = Require(options, "receipt");

            var profile = JsonSerializer.Deserialize<AcceptanceProfile>(
                File.ReadAllText(profilePath),
                new JsonSerializerOptions { PropertyNameCaseInsensitive = true })
                ?? throw new InvalidOperationException("Profile JSON was empty or invalid.");

            using var app = Application.Launch(exe);
            using var automation = new UIA3Automation();

            var window = app.GetMainWindow(automation, TimeSpan.FromSeconds(15))
                ?? throw new InvalidOperationException("No main window was discovered within 15 seconds.");

            if (!string.IsNullOrWhiteSpace(profile.WindowTitleContains) &&
                !window.Title.Contains(profile.WindowTitleContains, StringComparison.OrdinalIgnoreCase))
            {
                throw new InvalidOperationException(
                    $"Window title '{window.Title}' does not contain expected text '{profile.WindowTitleContains}'.");
            }

            var descendants = window.FindAllDescendants();
            var receipts = new List<ControlReceipt>();
            var failures = new List<string>();

            foreach (var expectation in profile.Controls)
            {
                if (!Enum.TryParse<ControlType>(expectation.ControlType, ignoreCase: true, out var expectedType))
                {
                    failures.Add($"{expectation.Name}: unknown ControlType '{expectation.ControlType}'.");
                    continue;
                }

                var candidates = descendants.Where(e =>
                    e.ControlType == expectedType &&
                    (string.IsNullOrEmpty(expectation.Name) || string.Equals(e.Name, expectation.Name, StringComparison.Ordinal)) &&
                    (string.IsNullOrEmpty(expectation.AutomationId) || string.Equals(e.AutomationId, expectation.AutomationId, StringComparison.Ordinal)))
                    .ToArray();

                if (candidates.Length == 0)
                {
                    failures.Add($"Missing {expectation.ControlType} '{expectation.Name}'.");
                    continue;
                }

                if (candidates.Length > 1)
                {
                    failures.Add($"Ambiguous {expectation.ControlType} '{expectation.Name}': {candidates.Length} matches.");
                    continue;
                }

                var element = candidates[0];
                var enabled = element.IsEnabled;
                var focusable = element.Properties.IsKeyboardFocusable.ValueOrDefault;

                if (expectation.MustBeEnabled && !enabled)
                {
                    failures.Add($"{expectation.Name}: expected enabled=true.");
                }

                if (expectation.MustBeKeyboardFocusable is bool expectedFocusable && focusable != expectedFocusable)
                {
                    failures.Add($"{expectation.Name}: expected keyboard focusable={expectedFocusable}, actual={focusable}.");
                }

                var supported = SupportedPatterns(element);
                foreach (var pattern in expectation.RequiredPatterns)
                {
                    if (!supported.Contains(pattern, StringComparer.OrdinalIgnoreCase))
                    {
                        failures.Add($"{expectation.Name}: missing required UIA pattern '{pattern}'.");
                    }
                }

                var exerciseResult = RunExercise(element, expectation, automation, failures);
                receipts.Add(new ControlReceipt(
                    element.Name,
                    element.AutomationId,
                    element.ControlType.ToString(),
                    enabled,
                    focusable,
                    supported,
                    expectation.Exercise,
                    exerciseResult));
            }

            var coreVersion = typeof(Application).Assembly.GetName().Version?.ToString() ?? "unknown";
            var uia3Version = typeof(UIA3Automation).Assembly.GetName().Version?.ToString() ?? "unknown";
            var overall = failures.Count == 0 ? "PASS" : "FAIL";

            var receipt = new AcceptanceReceipt(
                Path.GetFullPath(exe),
                DateTimeOffset.UtcNow,
                window.Title,
                coreVersion,
                uia3Version,
                receipts,
                overall);

            var json = JsonSerializer.Serialize(receipt, new JsonSerializerOptions { WriteIndented = true });
            Directory.CreateDirectory(Path.GetDirectoryName(Path.GetFullPath(receiptPath))!);
            File.WriteAllText(receiptPath, json + Environment.NewLine);

            Console.WriteLine($"ECO_UIA_ACCEPTANCE={overall}");
            Console.WriteLine($"ECO_UIA_RECEIPT={Path.GetFullPath(receiptPath)}");
            foreach (var failure in failures)
            {
                Console.Error.WriteLine("FAIL: " + failure);
            }

            return failures.Count == 0 ? 0 : 2;
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine("HARNESS_ERROR: " + ex);
            return 1;
        }
    }

    private static IReadOnlyList<string> SupportedPatterns(AutomationElement e)
    {
        var result = new List<string>();
        if (e.Patterns.Value.IsSupported) result.Add("Value");
        if (e.Patterns.Invoke.IsSupported) result.Add("Invoke");
        if (e.Patterns.Selection.IsSupported) result.Add("Selection");
        if (e.Patterns.SelectionItem.IsSupported) result.Add("SelectionItem");
        if (e.Patterns.Text.IsSupported) result.Add("Text");
        if (e.Patterns.Toggle.IsSupported) result.Add("Toggle");
        if (e.Patterns.ExpandCollapse.IsSupported) result.Add("ExpandCollapse");
        if (e.Patterns.Scroll.IsSupported) result.Add("Scroll");
        return result;
    }

    private static string RunExercise(
        AutomationElement element,
        ControlExpectation expectation,
        UIA3Automation automation,
        List<string> failures)
    {
        if (string.IsNullOrWhiteSpace(expectation.Exercise))
        {
            return "not requested";
        }

        try
        {
            switch (expectation.Exercise.Trim().ToLowerInvariant())
            {
                case "focus":
                    element.Focus();
                    Thread.Sleep(150);
                    var focused = automation.FocusedElement();
                    if (focused is null || !string.Equals(focused.Name, element.Name, StringComparison.Ordinal))
                    {
                        throw new InvalidOperationException($"focus landed on '{focused?.Name ?? "<none>"}'.");
                    }
                    return "focus confirmed";

                case "setvalue":
                    if (!element.Patterns.Value.TryGetPattern(out var valuePattern))
                    {
                        throw new InvalidOperationException("ValuePattern unavailable.");
                    }
                    var newValue = expectation.ExerciseValue ?? "ECO synthetic accessibility probe";
                    valuePattern.SetValue(newValue);
                    Thread.Sleep(100);
                    if (!string.Equals(valuePattern.Value.Value, newValue, StringComparison.Ordinal))
                    {
                        throw new InvalidOperationException("ValuePattern did not retain requested value.");
                    }
                    return "value set and read back";

                case "invoke":
                    if (!element.Patterns.Invoke.TryGetPattern(out var invokePattern))
                    {
                        throw new InvalidOperationException("InvokePattern unavailable.");
                    }
                    invokePattern.Invoke();
                    Thread.Sleep(100);
                    return "invoke completed";

                case "select":
                    var list = element.AsListBox();
                    var index = expectation.SelectIndex ?? 0;
                    var selected = list.Select(index);
                    Thread.Sleep(100);
                    if (list.SelectedItems.Length == 0 || !string.Equals(list.SelectedItems[0].Name, selected.Name, StringComparison.Ordinal))
                    {
                        throw new InvalidOperationException("Selection did not become observable through SelectionPattern.");
                    }
                    return $"selected index {index}";

                default:
                    throw new InvalidOperationException($"Unknown exercise '{expectation.Exercise}'.");
            }
        }
        catch (Exception ex)
        {
            failures.Add($"{expectation.Name}: exercise '{expectation.Exercise}' failed: {ex.Message}");
            return "failed: " + ex.Message;
        }
    }

    private static Dictionary<string, string> ParseArgs(string[] args)
    {
        var result = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase);
        for (var i = 0; i < args.Length; i++)
        {
            if (!args[i].StartsWith("--", StringComparison.Ordinal) || i + 1 >= args.Length)
            {
                throw new ArgumentException("Expected --key value arguments.");
            }
            result[args[i][2..]] = args[++i];
        }
        return result;
    }

    private static string Require(Dictionary<string, string> options, string key)
        => options.TryGetValue(key, out var value) && !string.IsNullOrWhiteSpace(value)
            ? value
            : throw new ArgumentException($"Missing --{key} value.");
}
