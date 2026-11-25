pragma ComponentBehavior: Bound

import Qt.labs.synchronizer
import Qt5Compat.GraphicalEffects
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Quickshell

import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions

Item { // Wrapper
    id: root

    readonly property string xdgConfigHome: Directories.config
    readonly property int typingDebounceInterval: 200
    readonly property int typingResultLimit: 15 // Should be enough to cover the whole view

    property string searchingText: LauncherSearch.query
    property bool showResults: searchingText != "" || ((LauncherSearch.results?.length ?? 0) > 0)
    readonly property real launcherWidth: 640
    implicitWidth: launcherWidth + Appearance.sizes.elevationMargin * 2
    implicitHeight: searchWidgetContent.implicitHeight + Appearance.sizes.elevationMargin * 2

    Component.onCompleted: {
        resultModel.values = LauncherSearch.results ?? [];
    }

    function focusFirstItem() {
        appResults.currentIndex = 0;
    }

    function focusSearchInput() {
        searchBar.forceFocus();
    }

    function disableExpandAnimation() {
        searchBar.animateWidth = false;
    }

    function cancelSearch() {
        searchBar.searchInput.selectAll();
        LauncherSearch.query = "";
        searchBar.animateWidth = true;
        resultModel.values = LauncherSearch.results ?? [];
        appResults.currentIndex = 0;
        appResults.contentY = 0;
        appResults.positionViewAtBeginning();
    }

    function setSearchingText(text) {
        searchBar.searchInput.text = text;
        LauncherSearch.query = text;
    }

    Keys.onPressed: event => {
        // Prevent Esc and Backspace from registering
        if (event.key === Qt.Key_Escape)
            return;

        // Handle Backspace: focus and delete character if not focused
        if (event.key === Qt.Key_Backspace) {
            if (!searchBar.searchInput.activeFocus) {
                root.focusSearchInput();
                if (event.modifiers & Qt.ControlModifier) {
                    // Delete word before cursor
                    let text = searchBar.searchInput.text;
                    let pos = searchBar.searchInput.cursorPosition;
                    if (pos > 0) {
                        // Find the start of the previous word
                        let left = text.slice(0, pos);
                        let match = left.match(/(\s*\S+)\s*$/);
                        let deleteLen = match ? match[0].length : 1;
                        searchBar.searchInput.text = text.slice(0, pos - deleteLen) + text.slice(pos);
                        searchBar.searchInput.cursorPosition = pos - deleteLen;
                    }
                } else {
                    // Delete character before cursor if any
                    if (searchBar.searchInput.cursorPosition > 0) {
                        searchBar.searchInput.text = searchBar.searchInput.text.slice(0, searchBar.searchInput.cursorPosition - 1) + searchBar.searchInput.text.slice(searchBar.searchInput.cursorPosition);
                        searchBar.searchInput.cursorPosition -= 1;
                    }
                }
                // Always move cursor to end after programmatic edit
                searchBar.searchInput.cursorPosition = searchBar.searchInput.text.length;
                event.accepted = true;
            }
            // If already focused, let TextField handle it
            return;
        }

        // Only handle visible printable characters (ignore control chars, arrows, etc.)
        if (event.text && event.text.length === 1 && event.key !== Qt.Key_Enter && event.key !== Qt.Key_Return && event.key !== Qt.Key_Delete && event.text.charCodeAt(0) >= 0x20)
        {
            if (!searchBar.searchInput.activeFocus) {
                root.focusSearchInput();
                // Insert the character at the cursor position
                searchBar.searchInput.text = searchBar.searchInput.text.slice(0, searchBar.searchInput.cursorPosition) + event.text + searchBar.searchInput.text.slice(searchBar.searchInput.cursorPosition);
                searchBar.searchInput.cursorPosition += 1;
                event.accepted = true;
                root.focusFirstItem();
            }
        }
    }

    StyledRectangularShadow {
        target: searchWidgetContent
    }

    Rectangle { // Background
        id: searchWidgetContent
        anchors {
            top: parent.top
            horizontalCenter: parent.horizontalCenter
            topMargin: Appearance.sizes.elevationMargin
        }
        clip: true
        implicitWidth: root.launcherWidth
        implicitHeight: columnLayout.implicitHeight
        radius: Appearance.rounding.windowRounding
        color: Appearance.colors.colLayer0
        border.width: 1
        border.color: Appearance.colors.colLayer0Border

        Behavior on implicitHeight {
            id: searchHeightBehavior
            enabled: GlobalStates.overviewOpen && root.showResults
            animation: Appearance.animation.elementMove.numberAnimation.createObject(this)
        }

        ColumnLayout {
            id: columnLayout
            width: root.launcherWidth
            spacing: 0

            layer.enabled: true
            layer.effect: OpacityMask {
                maskSource: Rectangle {
                    width: searchWidgetContent.width
                    height: searchWidgetContent.height
                    radius: searchWidgetContent.radius
                }
            }

            SearchBar {
                id: searchBar
                Layout.fillWidth: true
                Synchronizer on searchingText {
                    property alias source: root.searchingText
                }

                onAccepted: {
                    if (appResults.count > 0) {
                        let firstItem = appResults.itemAtIndex(0);
                        if (firstItem && firstItem.clicked) {
                            firstItem.clicked();
                        }
                    }
                }

                onNavigateDown: {
                    if (appResults.count > 0) {
                        appResults.forceActiveFocus();
                        appResults.currentIndex = 0;
                    }
                }
            }

            Rectangle {
                // Separator
                visible: root.showResults
                Layout.fillWidth: true
                height: 1
                color: Appearance.colors.colLayer0Border
            }

            RowLayout {
                visible: root.showResults && root.searchingText === ""
                Layout.fillWidth: true
                Layout.preferredHeight: 26
                Layout.leftMargin: 16
                Layout.rightMargin: 16
                Layout.topMargin: 4
                spacing: 8

                StyledText {
                    text: Translation.tr("INSTALLED APPLICATIONS")
                    font.pixelSize: 10
                    font.weight: Font.DemiBold
                    font.letterSpacing: 1.2
                    color: ColorUtils.transparentize(Appearance.colors.colSubtext, 0.3)
                }

                Item { Layout.fillWidth: true }

                StyledText {
                    text: (resultModel.values?.length ?? 0) + " " + Translation.tr("apps")
                    font.pixelSize: 10
                    font.weight: Font.Medium
                    color: ColorUtils.transparentize(Appearance.colors.colSubtext, 0.5)
                }
            }

            ListView { // App results
                id: appResults
                visible: root.showResults
                Layout.fillWidth: true
                implicitHeight: Math.min(480, appResults.contentHeight + topMargin + bottomMargin)
                clip: true
                topMargin: 8
                bottomMargin: 8
                leftMargin: 6
                rightMargin: 6
                spacing: 2
                KeyNavigation.up: searchBar.searchInput
                highlightMoveDuration: 100

                ScrollBar.vertical: ScrollBar {
                    id: resultScrollBar
                    policy: appResults.contentHeight > appResults.height ? ScrollBar.AlwaysOn : ScrollBar.AsNeeded
                    width: 6
                    active: true
                    contentItem: Rectangle {
                        implicitWidth: 4
                        radius: 2
                        color: resultScrollBar.hovered
                            ? Appearance.colors.colPrimary
                            : ColorUtils.transparentize(Appearance.colors.colSubtext, 0.6)
                    }
                }

                onFocusChanged: {
                    if (focus)
                        appResults.currentIndex = 0;
                }

                Connections {
                    target: root
                    function onSearchingTextChanged() {
                        if (appResults.count > 0)
                            appResults.currentIndex = 0;
                    }
                }

                Timer {
                    id: debounceTimer
                    interval: root.typingDebounceInterval
                    onTriggered: {
                        resultModel.values = LauncherSearch.results ?? [];
                    }
                }

                Connections {
                    target: LauncherSearch
                    function onResultsChanged() {
                        if (root.searchingText === "") {
                            resultModel.values = LauncherSearch.results ?? [];
                            appResults.currentIndex = 0;
                            appResults.contentY = 0;
                            appResults.positionViewAtBeginning();
                        } else {
                            resultModel.values = (LauncherSearch.results ?? []).slice(0, root.typingResultLimit);
                            debounceTimer.restart();
                        }
                        root.focusFirstItem();
                    }
                }

                model: ScriptModel {
                    id: resultModel
                    objectProp: "key"
                }

                delegate: SearchItem {
                    id: searchItem
                    required property var modelData
                    anchors.left: parent?.left
                    anchors.right: parent?.right
                    entry: modelData
                    query: StringUtils.cleanOnePrefix(root.searchingText, [Config.options.search.prefix.action, Config.options.search.prefix.app, Config.options.search.prefix.clipboard, Config.options.search.prefix.emojis, Config.options.search.prefix.math, Config.options.search.prefix.shellCommand, Config.options.search.prefix.webSearch])

                    Keys.onPressed: event => {
                        if (event.key === Qt.Key_Tab) {
                            if (LauncherSearch.results.length === 0)
                                return;
                            const tabbedText = searchItem.modelData.name;
                            LauncherSearch.query = tabbedText;
                            searchBar.searchInput.text = tabbedText;
                            event.accepted = true;
                            root.focusSearchInput();
                        }
                    }
                }
            }

            Rectangle {
                // Bottom separator
                visible: root.showResults
                Layout.fillWidth: true
                height: 1
                color: Appearance.colors.colLayer0Border
            }

            RowLayout {
                visible: root.showResults
                Layout.fillWidth: true
                Layout.preferredHeight: 34
                Layout.leftMargin: 16
                Layout.rightMargin: 16
                spacing: 12

                StyledText {
                    Layout.alignment: Qt.AlignVCenter
                    font.pixelSize: Appearance.font.pixelSize.smallest
                    color: Appearance.colors.colSubtext
                    text: {
                        let cnt = appResults.count;
                        if (root.searchingText === "") {
                            return cnt + (cnt === 1 ? " application" : " applications");
                        }
                        return cnt + (cnt === 1 ? " result" : " results");
                    }
                }

                Item { Layout.fillWidth: true }

                RowLayout {
                    Layout.alignment: Qt.AlignVCenter
                    spacing: 12

                    RowLayout {
                        spacing: 4
                        Rectangle {
                            implicitWidth: 18
                            implicitHeight: 18
                            radius: 4
                            color: Appearance.colors.colLayer1
                            border.width: 1
                            border.color: Appearance.colors.colLayer0Border
                            MaterialSymbol {
                                anchors.centerIn: parent
                                iconSize: 12
                                text: "keyboard_arrow_up"
                                color: Appearance.colors.colSubtext
                            }
                        }
                        Rectangle {
                            implicitWidth: 18
                            implicitHeight: 18
                            radius: 4
                            color: Appearance.colors.colLayer1
                            border.width: 1
                            border.color: Appearance.colors.colLayer0Border
                            MaterialSymbol {
                                anchors.centerIn: parent
                                iconSize: 12
                                text: "keyboard_arrow_down"
                                color: Appearance.colors.colSubtext
                            }
                        }
                        StyledText {
                            font.pixelSize: Appearance.font.pixelSize.smallest
                            color: Appearance.colors.colSubtext
                            text: "Navigate"
                        }
                    }

                    RowLayout {
                        spacing: 4
                        Rectangle {
                            implicitWidth: 20
                            implicitHeight: 18
                            radius: 4
                            color: Appearance.colors.colLayer1
                            border.width: 1
                            border.color: Appearance.colors.colLayer0Border
                            MaterialSymbol {
                                anchors.centerIn: parent
                                iconSize: 12
                                text: "keyboard_return"
                                color: Appearance.colors.colSubtext
                            }
                        }
                        StyledText {
                            font.pixelSize: Appearance.font.pixelSize.smallest
                            color: Appearance.colors.colSubtext
                            text: "Open"
                        }
                    }

                    RowLayout {
                        spacing: 4
                        Rectangle {
                            implicitWidth: 28
                            implicitHeight: 18
                            radius: 4
                            color: Appearance.colors.colLayer1
                            border.width: 1
                            border.color: Appearance.colors.colLayer0Border
                            StyledText {
                                anchors.centerIn: parent
                                font.pixelSize: 9
                                font.weight: Font.DemiBold
                                color: Appearance.colors.colSubtext
                                text: "ESC"
                            }
                        }
                        StyledText {
                            font.pixelSize: Appearance.font.pixelSize.smallest
                            color: Appearance.colors.colSubtext
                            text: "Close"
                        }
                    }
                }
            }
        }
    }
}
