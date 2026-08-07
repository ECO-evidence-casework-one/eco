import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

ApplicationWindow {
    id: root
    width: 1366
    height: 768
    minimumWidth: 1024
    minimumHeight: 640
    visible: true
    title: "ECO UI-0 — Qt framework gate"
    color: "#F4F7F6"

    property bool compact: width < 1180
    property var allActions: [
        "What warranty evidence do I have?",
        "Is written warranty confirmation missing?",
        "Show warranty-related communications",
        "Find warranty dates and promises",
        "What should I do next about the warranty?",
        "Build the Matter timeline",
        "Check for contradictions",
        "Explain why ECO raised this finding"
    ]
    property var displayedActions: allActions

    function updateSearch(q) {
        var n = q.toLowerCase().trim()
        if (n.length === 0) { displayedActions = allActions; return }
        var synonyms = n.indexOf("war") >= 0 ? " warranty guarantee written confirmation cover " : ""
        displayedActions = allActions.filter(function(v) {
            var h = v.toLowerCase()
            return h.indexOf(n) >= 0 || (synonyms.length > 0 && (h.indexOf("warranty") >= 0 || h.indexOf("confirmation") >= 0))
        })
    }

    header: ToolBar {
        height: 64
        background: Rectangle { color: "#123F3F" }
        RowLayout {
            anchors.fill: parent; anchors.leftMargin: 20; anchors.rightMargin: 20
            Label { text: "ECO"; color: "white"; font.pixelSize: 25; font.bold: true }
            Label { text: "Evidence & Casework One"; color: "#CBE5E1"; font.pixelSize: 14 }
            Item { Layout.fillWidth: true }
            Label { text: "UI-0 architecture proof • synthetic only"; color: "#CBE5E1" }
        }
    }

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Pane {
            Layout.preferredWidth: compact ? 72 : 208
            Layout.fillHeight: true
            padding: 12
            background: Rectangle { color: "#173F40" }
            ColumnLayout {
                width: parent.width; spacing: 8
                Repeater {
                    model: ["Workspace", "Evidence", "Case structure", "Work", "Reports"]
                    delegate: Button {
                        Layout.fillWidth: true
                        text: compact ? modelData.charAt(0) : modelData
                        flat: true
                        font.bold: modelData === "Workspace"
                        palette.buttonText: "white"
                        Accessible.name: modelData
                    }
                }
                Item { Layout.fillHeight: true }
                Button { Layout.fillWidth: true; text: compact ? "⚙" : "Settings"; flat: true; palette.buttonText: "white" }
            }
        }

        ScrollView {
            Layout.fillWidth: true
            Layout.fillHeight: true
            contentWidth: availableWidth
            clip: true

            ColumnLayout {
                width: parent.width
                spacing: 14
                anchors.margins: 22

                RowLayout {
                    Layout.fillWidth: true
                    ColumnLayout {
                        Layout.fillWidth: true
                        Label { text: "General Casework 1"; font.pixelSize: 26; font.bold: true; color: "#173B3C" }
                        Label { text: "Matter Workspace"; color: "#637878"; font.pixelSize: 14 }
                    }
                    Label { text: "10 evidence  •  9 readable  •  1 unresolved"; color: "#456565" }
                }

                Frame {
                    Layout.fillWidth: true
                    padding: 16
                    background: Rectangle { color: "white"; radius: 14; border.color: "#D3E1DE" }
                    ColumnLayout {
                        anchors.fill: parent; spacing: 10
                        Label { text: "What ECO found"; font.pixelSize: 18; font.bold: true; color: "#173B3C" }
                        Flow {
                            Layout.fillWidth: true; spacing: 8
                            Repeater {
                                model: ["4 proposed Facts", "1 person", "1 organisation", "1 communication", "1 issue", "1 missing item", "1 proposed task"]
                                delegate: Rectangle {
                                    width: chip.implicitWidth + 22; height: 32; radius: 16; color: "#E8F3F0"
                                    Label { id: chip; anchors.centerIn: parent; text: modelData; color: "#1D5A57" }
                                }
                            }
                        }
                    }
                }

                Frame {
                    Layout.fillWidth: true
                    padding: 16
                    background: Rectangle { color: "#FFFDF8"; radius: 14; border.color: "#E7D7AC" }
                    RowLayout {
                        anchors.fill: parent; spacing: 18
                        ColumnLayout {
                            Layout.fillWidth: true
                            Label { text: "Possible missing evidence"; font.pixelSize: 13; color: "#8B6A1F"; font.bold: true }
                            Label { text: "Written warranty confirmation"; font.pixelSize: 19; font.bold: true; color: "#352D1A" }
                            Label { Layout.fillWidth: true; wrapMode: Text.WordWrap; text: "The booking email says written confirmation will follow. ECO did not locate a separate confirmation in the readable evidence."; color: "#554D3B" }
                            Label { text: "Coverage: 9/10 evidence records machine-readable • 1 image unresolved"; color: "#7C725A"; font.pixelSize: 12 }
                        }
                        Button { text: "Why does ECO think this?" }
                        Button { text: "Review"; highlighted: true }
                    }
                }

                Frame {
                    Layout.fillWidth: true
                    padding: 16
                    background: Rectangle { color: "white"; radius: 14; border.color: "#D3E1DE" }
                    ColumnLayout {
                        anchors.fill: parent; spacing: 8
                        Label { text: "Actions & Questions"; font.pixelSize: 18; font.bold: true; color: "#173B3C" }
                        TextField {
                            id: search
                            Layout.fillWidth: true
                            placeholderText: "Search this Matter or ask ECO… Try ‘war’"
                            onTextChanged: root.updateSearch(text)
                            Accessible.name: "Search this Matter or ask ECO"
                        }
                        Repeater {
                            model: root.displayedActions
                            delegate: Button {
                                required property var modelData
                                Layout.fillWidth: true
                                text: modelData
                                flat: true
                                contentItem: Label { text: parent.text; color: "#174D4D"; elide: Text.ElideRight }
                            }
                        }
                    }
                }

                Frame {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 270
                    padding: 16
                    background: Rectangle { color: "white"; radius: 14; border.color: "#D3E1DE" }
                    ColumnLayout {
                        anchors.fill: parent; spacing: 8
                        RowLayout {
                            Layout.fillWidth: true
                            Label { text: "Ask ECO"; font.pixelSize: 18; font.bold: true; color: "#173B3C" }
                            Item { Layout.fillWidth: true }
                            Label { text: "Entire Matter • 10/10 considered • 1 unresolved"; color: "#627777"; font.pixelSize: 12 }
                        }
                        TextArea {
                            id: transcript
                            Layout.fillWidth: true; Layout.fillHeight: true
                            readOnly: true; selectByMouse: true; wrapMode: TextEdit.Wrap
                            text: "You\nWhat is known, uncertain and missing?\n\nECO\nKnown: the booking email records an engineer visit for 29 July 2026.\n\nPossible gap: written warranty confirmation may still be outstanding.\n\nWhy this might be wrong: confirmation may exist outside the imported evidence or in the unresolved image.\n\nSources: 01_repair_booking_email.eml • 04_warranty_schedule.txt\n\nThis text must remain normally selectable and copyable with Ctrl+C and the standard context menu."
                            Accessible.name: "AI conversation transcript"
                        }
                        RowLayout {
                            Layout.fillWidth: true
                            TextField { Layout.fillWidth: true; placeholderText: "Ask about this Matter…"; Accessible.name: "AI message" }
                            Button { text: "Send"; highlighted: true }
                        }
                    }
                }
                Item { Layout.preferredHeight: 20 }
            }
        }

        Pane {
            visible: !root.compact
            Layout.preferredWidth: 300
            Layout.fillHeight: true
            padding: 18
            background: Rectangle { color: "#EBF2F0"; border.color: "#D1DEDB" }
            ColumnLayout {
                width: parent.width; spacing: 12
                Label { text: "Current position"; font.pixelSize: 17; font.bold: true; color: "#173B3C" }
                Label { Layout.fillWidth: true; wrapMode: Text.WordWrap; text: "10 evidence records are linked. Four evidence-derived Facts await review. One possible warranty gap needs attention."; color: "#405A5A" }
                Rectangle { Layout.fillWidth: true; height: 1; color: "#CAD8D5" }
                Label { text: "Next recommended action"; font.bold: true; color: "#173B3C" }
                Label { Layout.fillWidth: true; wrapMode: Text.WordWrap; text: "Review the possible missing warranty confirmation and its source."; color: "#405A5A" }
                Button { Layout.fillWidth: true; text: "Review next finding"; highlighted: true }
                Item { Layout.fillHeight: true }
                Label { text: "UI-0 • framework comparison"; color: "#718281"; font.pixelSize: 11 }
            }
        }
    }
}
