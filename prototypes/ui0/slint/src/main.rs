use slint::{ModelRc, SharedString, VecModel};

slint::include_modules!();

fn filtered_actions(query: &str) -> Vec<SharedString> {
    let all = [
        "What warranty evidence do I have?",
        "Is written warranty confirmation missing?",
        "Show warranty-related communications",
        "Find warranty dates and promises",
        "What should I do next about the warranty?",
        "Build the Matter timeline",
        "Check for contradictions",
        "Explain why ECO raised this finding",
    ];
    let q = query.trim().to_lowercase();
    if q.is_empty() {
        return all.into_iter().map(SharedString::from).collect();
    }
    let warranty_intent = q.contains("war") || q.contains("guar") || q == "warranty";
    all.into_iter()
        .filter(|item| {
            let h = item.to_lowercase();
            h.contains(&q) || (warranty_intent && (h.contains("warranty") || h.contains("confirmation")))
        })
        .map(SharedString::from)
        .collect()
}

fn main() -> Result<(), slint::PlatformError> {
    std::env::set_var("SLINT_BACKEND", "winit-software");
    let ui = MainWindow::new()?;
    ui.set_actions(ModelRc::new(VecModel::from(filtered_actions(""))));

    let weak = ui.as_weak();
    ui.on_search_changed(move |query| {
        if let Some(ui) = weak.upgrade() {
            ui.set_actions(ModelRc::new(VecModel::from(filtered_actions(query.as_str()))));
        }
    });

    ui.run()
}
