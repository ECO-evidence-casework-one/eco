using Avalonia;
using Avalonia.Headless;
using Avalonia.Threading;
using Eco.Recovery;

namespace Eco.Recovery.Visuals;

internal static class Program
{
    [STAThread]
    private static int Main(string[] args)
    {
        var outputRoot = args.Length > 0
            ? Path.GetFullPath(args[0])
            : Path.GetFullPath("recovery-visuals");
        Directory.CreateDirectory(outputRoot);
        Environment.SetEnvironmentVariable("ECO_RECOVERY_STATE_ROOT", Path.Combine(outputRoot, "temporary-state"));

        AppBuilder.Configure<App>()
            .UseSkia()
            .UseHeadless(new AvaloniaHeadlessPlatformOptions
            {
                UseHeadlessDrawing = false
            })
            .SetupWithoutStarting();

        var captures = new (string FileName, double Width, double Height, Action<MainWindow> Configure)[]
        {
            ("01-guided-start.png", 1360, 840, window => window.OpenStartForVisuals()),
            ("02-current-position.png", 1360, 840, window => window.LoadSyntheticMatterForVisuals("CurrentPosition")),
            ("03-evidence.png", 1360, 840, window => window.LoadSyntheticMatterForVisuals("Evidence")),
            ("04-conversations.png", 1360, 840, window => window.LoadSyntheticMatterForVisuals("Conversations")),
            ("05-settings.png", 1360, 840, window => window.LoadSyntheticMatterForVisuals("Settings")),
            ("06-whats-new.png", 1360, 840, window => window.OpenWhatsNewForVisuals()),
            ("07-current-position-225-scale.png", 1600, 1000, window => window.LoadSyntheticMatterForVisuals("CurrentPosition", 2.25))
        };

        var records = new List<string>();
        foreach (var capture in captures)
        {
            var path = Path.Combine(outputRoot, capture.FileName);
            Capture(path, capture.Width, capture.Height, capture.Configure);
            var length = new FileInfo(path).Length;
            if (length < 10_000)
            {
                throw new InvalidOperationException($"Rendered frame is unexpectedly small: {capture.FileName} ({length} bytes)");
            }
            records.Add($"{capture.FileName}\t{capture.Width:0}x{capture.Height:0}\t{length}");
        }

        File.WriteAllLines(Path.Combine(outputRoot, "VISUAL_INVENTORY.tsv"), records);
        return 0;
    }

    private static void Capture(string path, double width, double height, Action<MainWindow> configure)
    {
        var window = new MainWindow
        {
            Width = width,
            Height = height,
            MinWidth = 1,
            MinHeight = 1
        };

        configure(window);
        window.Show();
        Dispatcher.UIThread.RunJobs();
        AvaloniaHeadlessPlatform.ForceRenderTimerTick();
        Dispatcher.UIThread.RunJobs();

        using var frame = window.CaptureRenderedFrame();
        frame.Save(path);

        window.Close();
        Dispatcher.UIThread.RunJobs();
    }
}
