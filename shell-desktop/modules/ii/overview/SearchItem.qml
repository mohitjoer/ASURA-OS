// pragma NativeMethodBehavior: AcceptThisObject
import qs
import qs.services
import qs.modules.common
import qs.modules.common.models
import qs.modules.common.widgets
import qs.modules.common.functions
import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Widgets
import Quickshell.Hyprland

RippleButton {
    id: root
    property LauncherSearchResult entry
    property string query
    property bool entryShown: entry?.shown ?? true
    property string itemType: entry?.type ?? Translation.tr("App")
    property string itemName: entry?.name ?? ""
    property var iconType: entry?.iconType
    property string iconName: entry?.iconName ?? ""
    property var itemExecute: entry?.execute
    property var fontType: switch(entry?.fontType) {
        case LauncherSearchResult.FontType.Monospace:
            return "monospace"
        case LauncherSearchResult.FontType.Normal:
            return "main"
        default:
            return "main"
    }
    property string itemClickActionName: entry?.verb ?? "Open"
    property string bigText: entry?.iconType === LauncherSearchResult.IconType.Text ? entry?.iconName ?? "" : ""
    property string materialSymbol: entry.iconType === LauncherSearchResult.IconType.Material ? entry?.iconName ?? "" : ""
    property string cliphistRawString: entry?.rawValue ?? ""
    property bool blurImage: entry?.blurImage ?? false

    property string itemSubtitle: {
        if (entry?.genericName && entry.genericName.length > 0 && entry.genericName.toLowerCase() !== root.itemName.toLowerCase()) {
            return entry.genericName;
        }
        if (entry?.comment && entry.comment.length > 0 && entry.comment.length < 60) {
            return entry.comment;
        }
        const name = (root.itemName || "").toLowerCase();
        if (name.includes("print")) return Translation.tr("Print & CUPS Management");
        if (name.includes("nvim") || name.includes("neovim")) return Translation.tr("Vim-Fork Terminal Code Editor");
        if (name.includes("opencode")) return Translation.tr("Terminal Development Environment");
        if (name.includes("resource monitor") || name.includes("btop")) return Translation.tr("System Resource Monitor & Task Manager");
        if (name.includes("thunar preferences")) return Translation.tr("File Manager Configuration");
        if (name.includes("thunar")) return Translation.tr("Fast File Manager");
        if (name.includes("removable drives")) return Translation.tr("Media & Storage Device Manager");
        if (name.includes("volume") || name.includes("pavucontrol")) return Translation.tr("Audio & Mixer Settings");
        if (name.includes("mpv")) return Translation.tr("High Performance Media Player");
        if (name.includes("toren vm")) return Translation.tr("Virtual Machine Manager");
        if (name.includes("toren store")) return Translation.tr("App & Package Store");
        if (name.includes("toren settings")) return Translation.tr("System Settings & Rice Hub");
        if (name.includes("install toren")) return Translation.tr("Live OS Installer");
        if (name.includes("network")) return Translation.tr("Network Configuration & VPN");
        if (root.itemType && root.itemType !== Translation.tr("App") && root.itemType !== "App") {
            return root.itemType;
        }
        return Translation.tr("Application");
    }
    
    visible: root.entryShown
    property int horizontalMargin: 4
    property int buttonHorizontalPadding: 10
    property int buttonVerticalPadding: 6
    property bool keyboardDown: false
    readonly property bool selected: (root.hovered || root.focus)

    implicitHeight: Math.max(50, rowLayout.implicitHeight + root.buttonVerticalPadding * 2)
    implicitWidth: rowLayout.implicitWidth + root.buttonHorizontalPadding * 2
    buttonRadius: 10
    colBackground: (root.down || root.keyboardDown) ? Appearance.colors.colPrimaryContainerActive : 
        (selected ? Appearance.colors.colPrimaryContainer : 
        ColorUtils.transparentize(Appearance.colors.colPrimaryContainer, 1))
    colBackgroundHover: Appearance.colors.colPrimaryContainer
    colRipple: Appearance.colors.colPrimaryContainerActive
    property color colForeground: selected ? Appearance.colors.colOnPrimaryContainer : Appearance.m3colors.m3onSurface

    readonly property string highlightPrefix: `<u><font color="${Appearance.colors.colPrimary}">`
    readonly property string highlightSuffix: `</font></u>`
    // Note that this highlighting is independent from the search
    // It's close, but does not accurately represent how the fuzzy algorithm works
    function highlightContent(content, query) {
        if (!query || query.length === 0 || content == query || fontType === "monospace")
            return StringUtils.escapeHtml(content);

        let contentLower = content.toLowerCase();
        let queryLower = query.toLowerCase();

        let result = "";
        let lastIndex = 0;
        let qIndex = 0;

        for (let i = 0; i < content.length && qIndex < query.length; i++) {
            if (contentLower[i] === queryLower[qIndex]) {
                // Add non-highlighted part (escaped)
                if (i > lastIndex)
                    result += StringUtils.escapeHtml(content.slice(lastIndex, i));
                // Add highlighted character (escaped)
                result += root.highlightPrefix + StringUtils.escapeHtml(content[i]) + root.highlightSuffix;
                lastIndex = i + 1;
                qIndex++;
            }
        }
        // Add the rest of the string (escaped)
        if (lastIndex < content.length)
            result += StringUtils.escapeHtml(content.slice(lastIndex));

        return result;
    }
    property string displayContent: highlightContent(root.itemName, root.query)

    property list<string> urls: {
        if (!root.itemName) return [];
        // Regular expression to match URLs
        const urlRegex = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi;
        const matches = root.itemName?.match(urlRegex)
            ?.filter(url => !url.includes("…")) // Elided = invalid
        return matches ? matches : [];
    }
    
    PointingHandInteraction {}

    background {
        anchors.fill: root
        anchors.leftMargin: root.horizontalMargin
        anchors.rightMargin: root.horizontalMargin
    }

    onClicked: {
        GlobalStates.overviewOpen = false
        root.itemExecute()
    }
    Keys.onPressed: (event) => {
        if (event.key === Qt.Key_Delete && event.modifiers === Qt.ShiftModifier) {
            const deleteAction = root.entry.actions.find(action => action.name == Translation.tr("Delete"));

            if (deleteAction) {
                deleteAction.execute()
            }
        } else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
            root.keyboardDown = true
            root.clicked()
            event.accepted = true;
        }
    }
    Keys.onReleased: (event) => {
        if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
            root.keyboardDown = false
            event.accepted = true;
        }
    }

    RowLayout {
        id: rowLayout
        spacing: 12
        anchors.fill: parent
        anchors.leftMargin: root.horizontalMargin + root.buttonHorizontalPadding
        anchors.rightMargin: root.horizontalMargin + root.buttonHorizontalPadding

        // Sleek Squircle Icon Frame for uniform alignment
        Rectangle {
            id: iconFrame
            Layout.alignment: Qt.AlignVCenter
            implicitWidth: 38
            implicitHeight: 38
            radius: 8
            color: root.selected
                ? ColorUtils.transparentize(Appearance.colors.colOnPrimaryContainer, 0.9)
                : Appearance.colors.colLayer1
            border.width: 1
            border.color: root.selected
                ? ColorUtils.transparentize(Appearance.colors.colOnPrimaryContainer, 0.8)
                : Appearance.colors.colLayer0Border

            Loader {
                id: iconLoader
                anchors.centerIn: parent
                active: true
                sourceComponent: switch(root.iconType) {
                    case LauncherSearchResult.IconType.Material:
                        return materialSymbolComponent
                    case LauncherSearchResult.IconType.Text:
                        return bigTextComponent
                    case LauncherSearchResult.IconType.System:
                        return iconImageComponent
                    case LauncherSearchResult.IconType.None:
                        return null
                    default:
                        return null
                }
            }
        }

        Component {
            id: iconImageComponent
            Item {
                implicitWidth: 26
                implicitHeight: 26

                readonly property string resolvedIconPath: {
                    if (!root.iconName || root.iconName.length === 0) return "";
                    if (root.iconName.startsWith("/") || root.iconName.startsWith("file://")) {
                        return root.iconName;
                    }

                    // Check local shell assets first
                    let localAsset = Quickshell.shellPath("assets/icons/" + root.iconName + ".svg");
                    if (localAsset && localAsset.length > 0) {
                        return "file://" + localAsset;
                    }

                    let guess = AppSearch.guessIcon(root.iconName);
                    if (guess && guess.length > 0 && guess !== root.iconName && guess !== "image-missing") {
                        let localGuess = Quickshell.shellPath("assets/icons/" + guess + ".svg");
                        if (localGuess && localGuess.length > 0) {
                            return "file://" + localGuess;
                        }
                    }

                    if (AppSearch.iconExists(root.iconName)) {
                        return Quickshell.iconPath(root.iconName);
                    }
                    if (guess && guess.length > 0 && guess !== root.iconName && guess !== "image-missing" && AppSearch.iconExists(guess)) {
                        return Quickshell.iconPath(guess);
                    }

                    const fallbacks = {
                        "computer": ["toren-vm", "virtualbox", "system-run", "utilities-terminal", "application-x-executable"],
                        "preferences-system": ["toren-settings", "toren-hub", "org.gnome.Settings", "system-run", "application-x-executable"],
                        "preferences-system-network": ["network-wired", "network-workgroup", "network-wireless", "preferences-system", "application-x-executable"],
                        "nm-connection-editor": ["preferences-system-network", "network-wired"],
                        "utilities-terminal": ["ghostty", "terminal", "x-terminal-emulator", "system-run"],
                        "utilities-system-monitor": ["btop", "system-run", "utilities-terminal"]
                    };
                    const list = fallbacks[root.iconName] || [];
                    for (let fb of list) {
                        let localFb = Quickshell.shellPath("assets/icons/" + fb + ".svg");
                        if (localFb && localFb.length > 0) {
                            return "file://" + localFb;
                        }
                        if (AppSearch.iconExists(fb)) {
                            return Quickshell.iconPath(fb);
                        }
                    }
                    if (AppSearch.iconExists("application-x-executable")) {
                        return Quickshell.iconPath("application-x-executable");
                    }
                    return "";
                }

                Image {
                    id: iconImg
                    anchors.fill: parent
                    source: parent.resolvedIconPath
                    sourceSize.width: 32
                    sourceSize.height: 32
                    visible: parent.resolvedIconPath.length > 0 && status === Image.Ready
                }

                MaterialSymbol {
                    anchors.centerIn: parent
                    visible: !iconImg.visible
                    iconSize: 22
                    color: root.colForeground
                    text: {
                        const name = (root.iconName || "").toLowerCase();
                        if (name.includes("vm") || name.includes("computer") || name.includes("virtual")) return "computer";
                        if (name.includes("setting") || name.includes("preferences") || name.includes("hub")) return "settings";
                        if (name.includes("network") || name.includes("wifi") || name.includes("connection")) return "lan";
                        if (name.includes("terminal") || name.includes("console") || name.includes("opencode")) return "terminal";
                        if (name.includes("monitor") || name.includes("btop") || name.includes("sysmon") || name.includes("resource")) return "speed";
                        if (name.includes("edit") || name.includes("text") || name.includes("nvim") || name.includes("zed") || name.includes("code")) return "edit_note";
                        if (name.includes("store") || name.includes("shop") || name.includes("market")) return "shopping_bag";
                        if (name.includes("audio") || name.includes("sound") || name.includes("music") || name.includes("volume")) return "volume_up";
                        if (name.includes("video") || name.includes("movie") || name.includes("mpv")) return "movie";
                        if (name.includes("image") || name.includes("photo") || name.includes("draw")) return "image";
                        if (name.includes("print")) return "print";
                        if (name.includes("usb") || name.includes("media") || name.includes("drive")) return "usb";
                        return "apps";
                    }
                }
            }
        }

        Component {
            id: materialSymbolComponent
            MaterialSymbol {
                text: root.materialSymbol
                iconSize: 22
                color: root.colForeground
            }
        }

        Component {
            id: bigTextComponent
            StyledText {
                text: root.bigText
                font.pixelSize: Appearance.font.pixelSize.larger
                color: root.colForeground
            }
        }

        // Main text with professional two-line hierarchy
        ColumnLayout {
            id: contentColumn
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignVCenter
            spacing: 2

            RowLayout {
                spacing: 6
                Layout.fillWidth: true

                Loader { // Checkmark for copied clipboard entry
                    visible: itemName == Quickshell.clipboardText && root.cliphistRawString
                    active: itemName == Quickshell.clipboardText && root.cliphistRawString
                    sourceComponent: Rectangle {
                        implicitWidth: 16
                        implicitHeight: 16
                        radius: 8
                        color: Appearance.colors.colPrimary
                        MaterialSymbol {
                            anchors.centerIn: parent
                            text: "check"
                            iconSize: 11
                            color: Appearance.m3colors.m3onPrimary
                        }
                    }
                }
                Repeater { // Favicons for links
                    model: root.query == root.itemName ? [] : root.urls
                    Favicon {
                        required property var modelData
                        size: 16
                        url: modelData
                    }
                }
                StyledText { // Item name/content
                    Layout.fillWidth: true
                    id: nameText
                    textFormat: Text.StyledText
                    font.pixelSize: Appearance.font.pixelSize.normal
                    font.weight: root.selected ? Font.DemiBold : Font.Normal
                    font.family: Appearance.font.family[root.fontType]
                    color: root.colForeground
                    horizontalAlignment: Text.AlignLeft
                    elide: Text.ElideRight
                    text: root.selected ? StringUtils.escapeHtml(root.itemName) : root.displayContent
                }
            }

            StyledText {
                font.pixelSize: Appearance.font.pixelSize.smallest
                color: root.selected
                    ? ColorUtils.transparentize(root.colForeground, 0.3)
                    : Appearance.colors.colSubtext
                visible: root.itemSubtitle.length > 0
                elide: Text.ElideRight
                Layout.fillWidth: true
                text: root.itemSubtitle
            }

            Loader { // Clipboard image preview
                active: root.cliphistRawString && Cliphist.entryIsImage(root.cliphistRawString)
                sourceComponent: CliphistImage {
                    Layout.fillWidth: true
                    entry: root.cliphistRawString
                    maxWidth: contentColumn.width
                    maxHeight: 140
                    blur: root.blurImage
                }
            }
        }

        // Action badge (visible on active item when no custom desktop actions)
        Rectangle {
            id: returnActionBadge
            visible: root.selected && ((root.entry?.actions?.length ?? 0) === 0)
            Layout.alignment: Qt.AlignVCenter
            implicitWidth: actionPillRow.implicitWidth + 14
            implicitHeight: 24
            radius: 6
            color: ColorUtils.transparentize(Appearance.colors.colPrimary, 0.85)
            border.width: 1
            border.color: ColorUtils.transparentize(Appearance.colors.colPrimary, 0.6)

            RowLayout {
                id: actionPillRow
                anchors.centerIn: parent
                spacing: 4

                MaterialSymbol {
                    iconSize: 13
                    color: Appearance.colors.colPrimary
                    text: "keyboard_return"
                }

                StyledText {
                    text: root.itemClickActionName || "Open"
                    font.pixelSize: Appearance.font.pixelSize.smallest
                    font.weight: Font.Medium
                    color: Appearance.colors.colPrimary
                }
            }
        }

        // Secondary actions (ONLY visible when item is selected/focused)
        RowLayout {
            visible: root.selected && ((root.entry?.actions?.length ?? 0) > 0)
            Layout.alignment: Qt.AlignVCenter
            spacing: 4
            Repeater {
                model: (root.entry?.actions ?? []).slice(0, 3)
                delegate: RippleButton {
                    id: actionButton
                    required property var modelData
                    property var iconType: modelData.iconType
                    property string iconName: modelData.iconName ?? ""
                    implicitHeight: 28
                    implicitWidth: 28
                    buttonRadius: 6

                    colBackground: Appearance.colors.colLayer1
                    colBackgroundHover: Appearance.colors.colSecondaryContainerHover
                    colRipple: Appearance.colors.colSecondaryContainerActive

                    contentItem: Item {
                        anchors.centerIn: parent
                        Loader {
                            anchors.centerIn: parent
                            active: actionButton.iconType === LauncherSearchResult.IconType.Material || actionButton.iconName === ""
                            sourceComponent: MaterialSymbol {
                                text: actionButton.iconName || "arrow_forward"
                                iconSize: 16
                                color: root.colForeground
                            }
                        }
                        Loader {
                            anchors.centerIn: parent
                            active: actionButton.iconType === LauncherSearchResult.IconType.System && actionButton.iconName !== ""
                            sourceComponent: Item {
                                implicitWidth: 16
                                implicitHeight: 16
                                Image {
                                    id: actImg
                                    anchors.fill: parent
                                    source: Quickshell.iconPath(actionButton.iconName)
                                    sourceSize.width: 16
                                    sourceSize.height: 16
                                    visible: source != "" && status === Image.Ready
                                }
                                MaterialSymbol {
                                    anchors.centerIn: parent
                                    visible: !actImg.visible
                                    text: "open_in_new"
                                    iconSize: 14
                                    color: root.colForeground
                                }
                            }
                        }
                    }

                    onClicked: modelData.execute()

                    StyledToolTip {
                        text: modelData.name
                    }
                }
            }
        }
    }
}
