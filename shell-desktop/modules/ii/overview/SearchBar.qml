pragma ComponentBehavior: Bound
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import Quickshell
import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions

RowLayout {
    id: root
    spacing: 12
    property bool animateWidth: false
    property alias searchInput: searchInput
    property string searchingText

    signal accepted()
    signal navigateDown()

    function forceFocus() {
        searchInput.forceActiveFocus();
    }

    enum SearchPrefixType { Action, App, Clipboard, Emojis, Math, ShellCommand, WebSearch, DefaultSearch }

    property var searchPrefixType: {
        if (root.searchingText.startsWith(Config.options.search.prefix.action)) return SearchBar.SearchPrefixType.Action;
        if (root.searchingText.startsWith(Config.options.search.prefix.app)) return SearchBar.SearchPrefixType.App;
        if (root.searchingText.startsWith(Config.options.search.prefix.clipboard)) return SearchBar.SearchPrefixType.Clipboard;
        if (root.searchingText.startsWith(Config.options.search.prefix.emojis)) return SearchBar.SearchPrefixType.Emojis;
        if (root.searchingText.startsWith(Config.options.search.prefix.math)) return SearchBar.SearchPrefixType.Math;
        if (root.searchingText.startsWith(Config.options.search.prefix.shellCommand)) return SearchBar.SearchPrefixType.ShellCommand;
        if (root.searchingText.startsWith(Config.options.search.prefix.webSearch)) return SearchBar.SearchPrefixType.WebSearch;
        return SearchBar.SearchPrefixType.DefaultSearch;
    }

    implicitHeight: 54
    Layout.fillWidth: true
    Layout.preferredHeight: 54
    Layout.leftMargin: 16
    Layout.rightMargin: 16

    Rectangle {
        id: iconContainer
        Layout.alignment: Qt.AlignVCenter
        implicitWidth: 32
        implicitHeight: 32
        radius: 8
        color: root.searchPrefixType !== SearchBar.SearchPrefixType.DefaultSearch
            ? ColorUtils.transparentize(Appearance.colors.colPrimary, 0.85)
            : Appearance.colors.colLayer1
        border.width: 1
        border.color: root.searchPrefixType !== SearchBar.SearchPrefixType.DefaultSearch
            ? ColorUtils.transparentize(Appearance.colors.colPrimary, 0.6)
            : Appearance.colors.colLayer0Border

        MaterialSymbol {
            anchors.centerIn: parent
            iconSize: 18
            color: root.searchPrefixType !== SearchBar.SearchPrefixType.DefaultSearch
                ? Appearance.colors.colPrimary
                : Appearance.colors.colSubtext
            text: switch (root.searchPrefixType) {
                case SearchBar.SearchPrefixType.Action: return "settings_suggest";
                case SearchBar.SearchPrefixType.App: return "apps";
                case SearchBar.SearchPrefixType.Clipboard: return "content_paste_search";
                case SearchBar.SearchPrefixType.Emojis: return "add_reaction";
                case SearchBar.SearchPrefixType.Math: return "calculate";
                case SearchBar.SearchPrefixType.ShellCommand: return "terminal";
                case SearchBar.SearchPrefixType.WebSearch: return "travel_explore";
                case SearchBar.SearchPrefixType.DefaultSearch: return "search";
                default: return "search";
            }
        }
    }

    TextField {
        id: searchInput
        Layout.fillWidth: true
        Layout.fillHeight: true
        Layout.alignment: Qt.AlignVCenter
        focus: GlobalStates.overviewOpen

        font {
            family: Appearance.font.family.main
            pixelSize: Appearance.font.pixelSize.larger
            hintingPreference: Font.PreferFullHinting
            variableAxes: Appearance.font.variableAxes.main
        }
        color: Appearance.colors.colOnLayer0
        placeholderTextColor: Appearance.colors.colSubtext
        placeholderText: switch (root.searchPrefixType) {
            case SearchBar.SearchPrefixType.Action: return Translation.tr("Search actions...");
            case SearchBar.SearchPrefixType.App: return Translation.tr("Search applications...");
            case SearchBar.SearchPrefixType.Clipboard: return Translation.tr("Search clipboard history...");
            case SearchBar.SearchPrefixType.Emojis: return Translation.tr("Search emojis...");
            case SearchBar.SearchPrefixType.Math: return Translation.tr("Calculate expression...");
            case SearchBar.SearchPrefixType.ShellCommand: return Translation.tr("Run shell command...");
            case SearchBar.SearchPrefixType.WebSearch: return Translation.tr("Search the web...");
            default: return Translation.tr("Search apps, files, or run commands...");
        }
        selectedTextColor: Appearance.colors.colOnSecondaryContainer
        selectionColor: Appearance.colors.colSecondaryContainer
        renderType: Text.NativeRendering
        verticalAlignment: TextInput.AlignVCenter
        leftPadding: 0
        rightPadding: 0
        topPadding: 0
        bottomPadding: 0

        background: null

        onTextChanged: LauncherSearch.query = text

        onAccepted: {
            root.accepted();
        }

        Keys.onPressed: event => {
            if (event.key === Qt.Key_Tab) {
                if (LauncherSearch.results.length === 0) return;
                const tabbedText = LauncherSearch.results[0].name;
                LauncherSearch.query = tabbedText;
                searchInput.text = tabbedText;
                event.accepted = true;
            } else if (event.key === Qt.Key_Down) {
                root.navigateDown();
                event.accepted = true;
            }
        }
    }

    RippleButton {
        id: clearButton
        visible: root.searchingText !== ""
        Layout.alignment: Qt.AlignVCenter
        implicitWidth: 28
        implicitHeight: 28
        buttonRadius: 14
        colBackground: hovered ? Appearance.colors.colLayer1Hover : "transparent"
        colBackgroundHover: Appearance.colors.colLayer1Hover
        colRipple: Appearance.colors.colLayer1Active

        PointingHandInteraction {}

        MaterialSymbol {
            anchors.centerIn: parent
            iconSize: 16
            color: clearButton.hovered ? Appearance.colors.colOnLayer0 : Appearance.colors.colSubtext
            text: "close"
        }

        onClicked: {
            searchInput.text = "";
            LauncherSearch.query = "";
            searchInput.forceActiveFocus();
        }
    }

    Rectangle {
        id: escBadge
        visible: root.searchingText === ""
        Layout.alignment: Qt.AlignVCenter
        implicitWidth: escText.implicitWidth + 14
        implicitHeight: 22
        radius: 6
        color: Appearance.colors.colLayer1
        border.width: 1
        border.color: Appearance.colors.colLayer0Border

        StyledText {
            id: escText
            anchors.centerIn: parent
            text: "ESC"
            font.pixelSize: Appearance.font.pixelSize.smallest
            font.weight: Font.DemiBold
            color: Appearance.colors.colSubtext
        }
    }
}
