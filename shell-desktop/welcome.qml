//@ pragma UseQApplication
//@ pragma Env QS_NO_RELOAD_POPUP=1
//@ pragma Env QT_QUICK_CONTROLS_STYLE=Basic
//@ pragma Env QT_QUICK_FLICKABLE_WHEEL_DECELERATION=10000
//@ pragma Env QT_SCALE_FACTOR=1

import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick.Window
import Quickshell
import Quickshell.Io
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions

ApplicationWindow {
    id: root

    property string firstRunFilePath: FileUtils.trimFileProtocol(`${Directories.state}/user/first_run.txt`)
    property string firstRunFileContent: "This file is just here to confirm you've been greeted :>"
    property real contentPadding: 12
    property bool showNextTime: false
    property int currentStep: 0

    readonly property var stepList: [
        { label: Translation.tr("Hello"), icon: "waving_hand" },
        { label: Translation.tr("Your Bar"), icon: "dock_to_bottom" },
        { label: Translation.tr("Apps & Dev"), icon: "apps" },
        { label: Translation.tr("Essentials"), icon: "keyboard" },
        { label: Translation.tr("Done"), icon: "check_circle" }
    ]
    readonly property int totalSteps: stepList.length

    visible: true
    title: Translation.tr("Toren Linux Welcome")

    minimumWidth: 840
    minimumHeight: 580
    width: 980
    height: 660
    color: Appearance.m3colors.m3background

    Component.onCompleted: {
        MaterialThemeLoader.reapplyTheme();
        Config.readWriteDelay = 0;
    }

    onClosing: {
        Quickshell.execDetached([
            "notify-send",
            Translation.tr("Welcome to Toren Linux"),
            Translation.tr("Enjoy! Reopen the welcome guide any time with <tt>Super+Shift+Alt+/</tt>. Open Settings with <tt>Super+I</tt>"),
            "-a", "Shell"
        ]);
        Qt.quit();
    }

    function advance() {
        if (currentStep < totalSteps - 1) {
            currentStep++;
        } else {
            finish();
        }
    }

    function back() {
        if (currentStep > 0) {
            currentStep--;
        }
    }

    function finish() {
        if (!showNextTime) {
            Quickshell.execDetached(["bash", "-c", `echo '${StringUtils.shellSingleQuoteEscape(root.firstRunFileContent)}' > '${StringUtils.shellSingleQuoteEscape(root.firstRunFilePath)}'`]);
        }
        root.close();
    }

    function runCmd(argv) {
        if (!argv || argv.length === 0) return;
        Quickshell.execDetached(argv);
    }

    component AppToolCard: Rectangle {
        id: cardRoot
        property string iconName: "apps"
        property string titleText: ""
        property string badgeText: ""
        property string descText: ""
        property var command: []

        Layout.fillWidth: true
        implicitHeight: 46
        radius: 8
        color: cardMouse.containsMouse ? Appearance.m3colors.m3surfaceContainerHigh : Appearance.m3colors.m3surfaceContainerLowest
        border.width: 1
        border.color: cardMouse.containsMouse ? Appearance.m3colors.m3primary : Appearance.colors.colOutlineVariant

        Behavior on color { ColorAnimation { duration: 120 } }
        Behavior on border.color { ColorAnimation { duration: 120 } }

        RowLayout {
            anchors.fill: parent
            anchors.leftMargin: 10
            anchors.rightMargin: 10
            spacing: 10

            Rectangle {
                width: 28
                height: 28
                radius: 6
                color: Appearance.m3colors.m3surfaceContainerHigh

                MaterialSymbol {
                    anchors.centerIn: parent
                    text: cardRoot.iconName
                    iconSize: 16
                    color: Appearance.m3colors.m3primary
                }
            }

            ColumnLayout {
                Layout.fillWidth: true
                spacing: 1

                RowLayout {
                    spacing: 6
                    StyledText {
                        text: cardRoot.titleText
                        font.weight: Font.DemiBold
                        font.pixelSize: 13
                        color: Appearance.colors.colOnLayer0
                    }
                    Rectangle {
                        visible: cardRoot.badgeText.length > 0
                        implicitWidth: badgeLabel.implicitWidth + 8
                        implicitHeight: 16
                        radius: 4
                        color: Appearance.m3colors.m3surfaceContainerHigh
                        StyledText {
                            id: badgeLabel
                            anchors.centerIn: parent
                            text: cardRoot.badgeText
                            font.pixelSize: 9
                            font.weight: Font.Bold
                            color: Appearance.m3colors.m3primary
                        }
                    }
                }

                StyledText {
                    text: cardRoot.descText
                    font.pixelSize: Appearance.font.pixelSize.smaller
                    color: Appearance.colors.colSubtext
                    elide: Text.ElideRight
                    Layout.fillWidth: true
                }
            }

            Rectangle {
                visible: cardRoot.command && cardRoot.command.length > 0
                width: 24
                height: 24
                radius: 12
                color: cardMouse.containsMouse ? Appearance.m3colors.m3primary : "transparent"

                MaterialSymbol {
                    anchors.centerIn: parent
                    text: "open_in_new"
                    iconSize: 13
                    color: cardMouse.containsMouse ? Appearance.m3colors.m3onPrimary : Appearance.colors.colSubtext
                }
            }
        }

        MouseArea {
            id: cardMouse
            anchors.fill: parent
            hoverEnabled: true
            cursorShape: (cardRoot.command && cardRoot.command.length > 0) ? Qt.PointingHandCursor : Qt.ArrowCursor
            onClicked: {
                if (cardRoot.command && cardRoot.command.length > 0) {
                    root.runCmd(cardRoot.command);
                }
            }
        }
    }

    component ShortcutCard: Rectangle {
        id: cardRoot
        property string comboText: ""
        property string titleText: ""
        property string descText: ""
        property var command: []

        Layout.fillWidth: true
        implicitHeight: 52
        radius: 8
        color: cardMouse.containsMouse ? Appearance.m3colors.m3surfaceContainerHigh : Appearance.m3colors.m3surfaceContainerLowest
        border.width: 1
        border.color: cardMouse.containsMouse ? Appearance.m3colors.m3primary : Appearance.colors.colOutlineVariant

        Behavior on color { ColorAnimation { duration: 120 } }
        Behavior on border.color { ColorAnimation { duration: 120 } }

        RowLayout {
            anchors.fill: parent
            anchors.leftMargin: 12
            anchors.rightMargin: 12
            spacing: 8

            // Left Keycap
            Rectangle {
                width: 94
                height: 26
                radius: 4
                color: Appearance.m3colors.m3surfaceContainerHigh
                Layout.alignment: Qt.AlignVCenter

                StyledText {
                    anchors.centerIn: parent
                    text: cardRoot.comboText
                    font.pixelSize: 11
                    font.weight: Font.Bold
                    color: Appearance.colors.colOnLayer0
                }
            }

            // Centered Text Block
            ColumnLayout {
                Layout.fillWidth: true
                Layout.alignment: Qt.AlignVCenter
                spacing: 2

                StyledText {
                    Layout.alignment: Qt.AlignHCenter
                    horizontalAlignment: Text.AlignHCenter
                    text: cardRoot.titleText
                    font.weight: Font.DemiBold
                    font.pixelSize: 13
                    color: Appearance.colors.colOnLayer0
                }

                StyledText {
                    Layout.alignment: Qt.AlignHCenter
                    horizontalAlignment: Text.AlignHCenter
                    text: cardRoot.descText
                    font.pixelSize: 11
                    color: Appearance.colors.colSubtext
                    elide: Text.ElideRight
                }
            }

            // Right Spacer balancing the keycap to guarantee exact card-center alignment
            Item {
                width: 94
                height: 26
                Layout.alignment: Qt.AlignVCenter

                Rectangle {
                    anchors.right: parent.right
                    anchors.verticalCenter: parent.verticalCenter
                    width: 24
                    height: 24
                    radius: 12
                    color: cardMouse.containsMouse ? Appearance.m3colors.m3primary : "transparent"

                    MaterialSymbol {
                        anchors.centerIn: parent
                        text: "open_in_new"
                        iconSize: 13
                        color: cardMouse.containsMouse ? Appearance.m3colors.m3onPrimary : "transparent"
                    }
                }
            }
        }

        MouseArea {
            id: cardMouse
            anchors.fill: parent
            hoverEnabled: true
            cursorShape: (cardRoot.command && cardRoot.command.length > 0) ? Qt.PointingHandCursor : Qt.ArrowCursor
            onClicked: {
                if (cardRoot.command && cardRoot.command.length > 0) {
                    root.runCmd(cardRoot.command);
                }
            }
        }
    }

    // Keyboard shortcuts to navigate
    Shortcut { sequence: "Right"; onActivated: root.advance() }
    Shortcut { sequence: "Left"; onActivated: root.back() }
    Shortcut { sequence: "Return"; onActivated: root.advance() }
    Shortcut { sequence: "Escape"; onActivated: root.close() }

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: contentPadding
        spacing: 10

        // ==========================================
        // 1. TOP HEADER & STEP INDICATORS
        // ==========================================
        Rectangle {
            Layout.fillWidth: true
            implicitHeight: 52
            color: Appearance.m3colors.m3surfaceContainerLow
            radius: Appearance.rounding.windowRounding - 2

            RowLayout {
                anchors.fill: parent
                anchors.leftMargin: 16
                anchors.rightMargin: 16
                spacing: 12

                // Brand Mark
                RowLayout {
                    spacing: 6
                    Layout.alignment: Qt.AlignVCenter

                    Text {
                        text: "//"
                        color: "#e2342a"
                        font.family: Appearance.font.family.title
                        font.pixelSize: 18
                        font.weight: Font.Black
                    }

                    StyledText {
                        text: "TOREN"
                        font.pixelSize: Appearance.font.pixelSize.medium
                        font.weight: Font.Bold
                        color: Appearance.colors.colOnLayer0
                    }

                    Rectangle {
                        width: 1
                        height: 14
                        color: Appearance.colors.colOutline
                        Layout.margins: 4
                    }

                    StyledText {
                        text: Translation.tr("FIRST RUN")
                        font.pixelSize: Appearance.font.pixelSize.smaller
                        font.letterSpacing: 1.2
                        color: Appearance.colors.colSubtext
                    }
                }

                Item { Layout.fillWidth: true }

                // Step Progression Tabs
                RowLayout {
                    spacing: 6
                    Layout.alignment: Qt.AlignVCenter

                    Repeater {
                        model: root.stepList

                        Rectangle {
                            id: stepTab
                            required property var modelData
                            required property int index

                            implicitWidth: tabRow.implicitWidth + 18
                            implicitHeight: 32
                            radius: Appearance.rounding.full
                            color: root.currentStep === index 
                                   ? Appearance.m3colors.m3primary 
                                   : (stepMouse.containsMouse ? Appearance.m3colors.m3surfaceContainerHigh : "transparent")
                            border.width: root.currentStep === index ? 0 : 1
                            border.color: root.currentStep > index ? Appearance.m3colors.m3primary : Appearance.colors.colOutlineVariant

                            Behavior on color { ColorAnimation { duration: 150 } }

                            RowLayout {
                                id: tabRow
                                anchors.centerIn: parent
                                spacing: 6

                                StyledText {
                                    text: "0" + (index + 1)
                                    font.pixelSize: Appearance.font.pixelSize.smaller
                                    font.weight: Font.Bold
                                    color: root.currentStep === index 
                                           ? Appearance.m3colors.m3onPrimary 
                                           : (root.currentStep > index ? Appearance.m3colors.m3primary : Appearance.colors.colSubtext)
                                }

                                StyledText {
                                    text: modelData.label
                                    font.pixelSize: Appearance.font.pixelSize.small
                                    font.weight: root.currentStep === index ? Font.Bold : Font.Normal
                                    color: root.currentStep === index 
                                           ? Appearance.m3colors.m3onPrimary 
                                           : (root.currentStep > index ? Appearance.colors.colOnLayer0 : Appearance.colors.colSubtext)
                                }
                            }

                            MouseArea {
                                id: stepMouse
                                anchors.fill: parent
                                hoverEnabled: true
                                cursorShape: Qt.PointingHandCursor
                                onClicked: root.currentStep = index
                            }
                        }
                    }
                }

                Item { Layout.fillWidth: true }

                // Window Controls
                RowLayout {
                    spacing: 8
                    Layout.alignment: Qt.AlignVCenter

                    StyledText {
                        font.pixelSize: Appearance.font.pixelSize.smaller
                        text: Translation.tr("Show next time")
                        color: Appearance.colors.colSubtext
                    }

                    StyledSwitch {
                        checked: root.showNextTime
                        scale: 0.65
                        Layout.alignment: Qt.AlignVCenter
                        onCheckedChanged: {
                            root.showNextTime = checked;
                            if (checked) {
                                Quickshell.execDetached(["rm", "-f", root.firstRunFilePath]);
                            } else {
                                Quickshell.execDetached(["bash", "-c", `echo '${StringUtils.shellSingleQuoteEscape(root.firstRunFileContent)}' > '${StringUtils.shellSingleQuoteEscape(root.firstRunFilePath)}'`]);
                            }
                        }
                    }

                    RippleButton {
                        buttonRadius: Appearance.rounding.full
                        implicitWidth: 32
                        implicitHeight: 32
                        onClicked: root.close()
                        contentItem: MaterialSymbol {
                            anchors.centerIn: parent
                            text: "close"
                            iconSize: 18
                            color: Appearance.colors.colOnLayer0
                        }
                    }
                }
            }
        }

        // ==========================================
        // 2. MAIN CONTENT STAGE
        // ==========================================
        Rectangle {
            Layout.fillWidth: true
            Layout.fillHeight: true
            color: Appearance.m3colors.m3surfaceContainerLow
            radius: Appearance.rounding.windowRounding - 2
            clip: true

            StackLayout {
                id: stackView
                anchors.fill: parent
                anchors.margins: 24
                currentIndex: root.currentStep

                // ------------------------------------------
                // PAGE 0: HELLO
                // ------------------------------------------
                Item {
                    id: pageHello

                    RowLayout {
                        anchors.fill: parent
                        spacing: 32

                        // Left Hero Panel
                        Rectangle {
                            Layout.fillHeight: true
                            Layout.preferredWidth: 320
                            radius: Appearance.rounding.cardRounding
                            color: Appearance.m3colors.m3surfaceContainerLowest
                            border.width: 1
                            border.color: Appearance.colors.colOutlineVariant
                            clip: true

                            ColumnLayout {
                                anchors.fill: parent
                                anchors.margins: 24
                                spacing: 16

                                Item { Layout.fillHeight: true }

                                RowLayout {
                                    spacing: 8
                                    Text {
                                        text: "//"
                                        color: "#e2342a"
                                        font.family: Appearance.font.family.title
                                        font.pixelSize: 38
                                        font.weight: Font.Black
                                    }
                                    StyledText {
                                        text: "TOREN"
                                        font.pixelSize: 28
                                        font.weight: Font.Bold
                                        color: Appearance.colors.colOnLayer0
                                    }
                                }

                                StyledText {
                                    text: Translation.tr("PURE MONOCHROME DESKTOP")
                                    font.pixelSize: Appearance.font.pixelSize.smaller
                                    font.letterSpacing: 2.0
                                    font.weight: Font.DemiBold
                                    color: "#e2342a"
                                }

                                StyledText {
                                    text: Translation.tr("Engineered for minimal friction, high efficiency, and clean ergonomics on Hyprland.")
                                    font.pixelSize: Appearance.font.pixelSize.small
                                    wrapMode: Text.WordWrap
                                    Layout.fillWidth: true
                                    color: Appearance.colors.colSubtext
                                    lineHeight: 1.35
                                }

                                Rectangle {
                                    Layout.fillWidth: true
                                    height: 1
                                    color: Appearance.colors.colOutlineVariant
                                }

                                RowLayout {
                                    spacing: 6
                                    MaterialSymbol {
                                        text: "verified"
                                        iconSize: 16
                                        color: Appearance.m3colors.m3primary
                                    }
                                    StyledText {
                                        text: "v0.1-pre-alpha · Void Linux"
                                        font.pixelSize: Appearance.font.pixelSize.smaller
                                        color: Appearance.colors.colSubtext
                                    }
                                }

                                Item { Layout.fillHeight: true }
                            }
                        }

                        // Right Intro Panel
                        ColumnLayout {
                            Layout.fillWidth: true
                            Layout.fillHeight: true
                            spacing: 18

                            Item { Layout.fillHeight: true }

                            StyledText {
                                text: Translation.tr("Welcome to Toren Linux")
                                font.pixelSize: 32
                                font.weight: Font.Bold
                                color: Appearance.colors.colOnLayer0
                            }

                            StyledText {
                                text: Translation.tr("Your desktop is pre-configured and dressed. This guided tour sets up your bar position, introduces bundled apps and developer tools, and covers essential shortcuts.")
                                font.pixelSize: Appearance.font.pixelSize.medium
                                color: Appearance.colors.colSubtext
                                wrapMode: Text.WordWrap
                                Layout.fillWidth: true
                                lineHeight: 1.4
                            }

                            Rectangle {
                                Layout.fillWidth: true
                                height: 1
                                color: Appearance.colors.colOutlineVariant
                            }

                            // Highlights
                            ColumnLayout {
                                spacing: 14
                                Layout.fillWidth: true

                                RowLayout {
                                    spacing: 12
                                    Rectangle {
                                        width: 36
                                        height: 36
                                        radius: 18
                                        color: Appearance.m3colors.m3surfaceContainerHigh
                                        MaterialSymbol {
                                            anchors.centerIn: parent
                                            text: "dock_to_bottom"
                                            iconSize: 20
                                            color: Appearance.m3colors.m3primary
                                        }
                                    }
                                    ColumnLayout {
                                        spacing: 2
                                        StyledText {
                                            text: Translation.tr("Status Bar & Edges")
                                            font.weight: Font.DemiBold
                                            color: Appearance.colors.colOnLayer0
                                        }
                                        StyledText {
                                            text: Translation.tr("Choose whether your bar sits at the top, bottom, or sides.")
                                            font.pixelSize: Appearance.font.pixelSize.small
                                            color: Appearance.colors.colSubtext
                                        }
                                    }
                                }

                                RowLayout {
                                    spacing: 12
                                    Rectangle {
                                        width: 36
                                        height: 36
                                        radius: 18
                                        color: Appearance.m3colors.m3surfaceContainerHigh
                                        MaterialSymbol {
                                            anchors.centerIn: parent
                                            text: "apps"
                                            iconSize: 20
                                            color: Appearance.m3colors.m3primary
                                        }
                                    }
                                    ColumnLayout {
                                        spacing: 2
                                        StyledText {
                                            text: Translation.tr("Apps & Dev Stack")
                                            font.weight: Font.DemiBold
                                            color: Appearance.colors.colOnLayer0
                                        }
                                        StyledText {
                                            text: Translation.tr("Explore the pre-installed applications and developer environment.")
                                            font.pixelSize: Appearance.font.pixelSize.small
                                            color: Appearance.colors.colSubtext
                                        }
                                    }
                                }

                                RowLayout {
                                    spacing: 12
                                    Rectangle {
                                        width: 36
                                        height: 36
                                        radius: 18
                                        color: Appearance.m3colors.m3surfaceContainerHigh
                                        MaterialSymbol {
                                            anchors.centerIn: parent
                                            text: "keyboard"
                                            iconSize: 20
                                            color: Appearance.m3colors.m3primary
                                        }
                                    }
                                    ColumnLayout {
                                        spacing: 2
                                        StyledText {
                                            text: Translation.tr("Essential Shortcuts")
                                            font.weight: Font.DemiBold
                                            color: Appearance.colors.colOnLayer0
                                        }
                                        StyledText {
                                            text: Translation.tr("Learn the 8 key combinations that run your entire workspace.")
                                            font.pixelSize: Appearance.font.pixelSize.small
                                            color: Appearance.colors.colSubtext
                                        }
                                    }
                                }
                            }

                            Rectangle {
                                id: ghCard
                                Layout.fillWidth: true
                                visible: !ghDismissed
                                radius: Appearance.rounding.normal
                                color: Appearance.m3colors.m3surfaceContainerHigh
                                implicitHeight: ghLoginRow.implicitHeight + 24

                                property bool ghLoggedIn: false
                                property bool ghDismissed: false

                                Process {
                                    id: ghStatusProc
                                    running: true
                                    command: ["gh", "auth", "status"]
                                    onExited: (exitCode) => {
                                        ghCard.ghLoggedIn = (exitCode === 0);
                                    }
                                }

                                RowLayout {
                                    id: ghLoginRow
                                    anchors.fill: parent
                                    anchors.margins: 12
                                    spacing: 12

                                    MaterialSymbol {
                                        text: ghCard.ghLoggedIn ? "check_circle" : "passkey"
                                        iconSize: 22
                                        color: ghCard.ghLoggedIn ? Appearance.m3colors.m3tertiary : Appearance.m3colors.m3primary
                                    }

                                    ColumnLayout {
                                        Layout.fillWidth: true
                                        spacing: 2
                                        StyledText {
                                            text: ghCard.ghLoggedIn
                                                ? Translation.tr("GitHub Connected")
                                                : Translation.tr("Developer Setup (Optional): GitHub Login")
                                            font.weight: Font.DemiBold
                                            color: Appearance.colors.colOnLayer0
                                        }
                                        StyledText {
                                            text: ghCard.ghLoggedIn
                                                ? Translation.tr("GitHub CLI is signed in and ready for your coding agents.")
                                                : Translation.tr("Developers: sign in to GitHub first for PRs and coding agents. Everyone else can skip this.")
                                            font.pixelSize: Appearance.font.pixelSize.small
                                            color: Appearance.colors.colSubtext
                                            wrapMode: Text.WordWrap
                                            Layout.fillWidth: true
                                        }
                                    }

                                    RowLayout {
                                        visible: !ghCard.ghLoggedIn
                                        spacing: 8

                                        RippleButton {
                                            buttonText: Translation.tr("Skip")
                                            colBackground: Appearance.m3colors.m3surfaceContainerHighest
                                            colText: Appearance.colors.colOnLayer0
                                            onClicked: ghCard.ghDismissed = true
                                        }

                                        RippleButton {
                                            buttonText: Translation.tr("Sign in to GitHub")
                                            colBackground: Appearance.m3colors.m3primary
                                            colText: Appearance.m3colors.m3onPrimary
                                            onClicked: {
                                                Quickshell.execDetached(["toren-app", "terminal", "--", "gh", "auth", "login"]);
                                            }
                                        }
                                    }
                                }
                            }

                            Item { Layout.fillHeight: true }
                        }
                    }
                }

                // ------------------------------------------
                // PAGE 1: YOUR BAR
                // ------------------------------------------
                Item {
                    id: pageBar

                    ColumnLayout {
                        anchors.fill: parent
                        spacing: 16

                        ColumnLayout {
                            spacing: 4
                            StyledText {
                                text: Translation.tr("Choose your edge")
                                font.pixelSize: 24
                                font.weight: Font.Bold
                                color: Appearance.colors.colOnLayer0
                            }
                            StyledText {
                                text: Translation.tr("Select where your status bar sits and how its silhouette is shaped. Changes apply live.")
                                font.pixelSize: Appearance.font.pixelSize.small
                                color: Appearance.colors.colSubtext
                            }
                        }

                        Rectangle {
                            Layout.fillWidth: true
                            height: 1
                            color: Appearance.colors.colOutlineVariant
                        }

                        RowLayout {
                            Layout.fillWidth: true
                            Layout.fillHeight: true
                            spacing: 24

                            // Left Controls
                            ColumnLayout {
                                Layout.preferredWidth: 380
                                Layout.fillHeight: true
                                spacing: 16

                                ContentSubsection {
                                    title: Translation.tr("Bar position")
                                    ConfigSelectionArray {
                                        currentValue: (Config.options.bar.bottom ? 1 : 0) | (Config.options.bar.vertical ? 2 : 0)
                                        onSelected: newValue => {
                                            Config.options.bar.bottom = (newValue & 1) !== 0;
                                            Config.options.bar.vertical = (newValue & 2) !== 0;
                                        }
                                        options: [
                                            { displayName: Translation.tr("Top"), icon: "arrow_upward", value: 0 },
                                            { displayName: Translation.tr("Left"), icon: "arrow_back", value: 2 },
                                            { displayName: Translation.tr("Bottom"), icon: "arrow_downward", value: 1 },
                                            { displayName: Translation.tr("Right"), icon: "arrow_forward", value: 3 }
                                        ]
                                    }
                                }

                                ContentSubsection {
                                    title: Translation.tr("Bar silhouette & corners")
                                    ConfigSelectionArray {
                                        currentValue: Config.options.bar.cornerStyle
                                        onSelected: newValue => {
                                            Config.options.bar.cornerStyle = newValue;
                                        }
                                        options: [
                                            { displayName: Translation.tr("Hug"), icon: "line_curve", value: 0 },
                                            { displayName: Translation.tr("Float"), icon: "page_header", value: 1 },
                                            { displayName: Translation.tr("Rect"), icon: "toolbar", value: 2 }
                                        ]
                                    }
                                }

                                Item { Layout.fillHeight: true }

                                NoticeBox {
                                    Layout.fillWidth: true
                                    text: Translation.tr("Tip: You can customize panel widgets, sound, and CPU monitors anytime in Toren Settings.")
                                }
                            }

                            // Right Visual Preview
                            Rectangle {
                                Layout.fillWidth: true
                                Layout.fillHeight: true
                                radius: Appearance.rounding.cardRounding
                                color: Appearance.m3colors.m3surfaceContainerLowest
                                border.width: 1
                                border.color: Appearance.colors.colOutlineVariant
                                clip: true

                                ColumnLayout {
                                    anchors.centerIn: parent
                                    spacing: 14

                                    // Monitor Mockup
                                    Rectangle {
                                        Layout.alignment: Qt.AlignHCenter
                                        width: 300
                                        height: 190
                                        radius: 10
                                        color: "#0d0d12"
                                        border.width: 2
                                        border.color: Appearance.colors.colOutlineVariant

                                        // Screen Display Area
                                        Item {
                                            anchors.fill: parent
                                            anchors.margins: 4
                                            clip: true

                                            // Mockup Bar
                                            Rectangle {
                                                id: previewBarMock
                                                color: Appearance.m3colors.m3primary
                                                radius: Config.options.bar.cornerStyle === 1 ? 4 : (Config.options.bar.cornerStyle === 0 ? 0 : 0)

                                                // Dynamic positioning based on user choice
                                                anchors.top: !Config.options.bar.bottom && !Config.options.bar.vertical ? parent.top : undefined
                                                anchors.bottom: Config.options.bar.bottom && !Config.options.bar.vertical ? parent.bottom : undefined
                                                anchors.left: Config.options.bar.vertical && !Config.options.bar.bottom ? parent.left : (Config.options.bar.vertical ? undefined : parent.left)
                                                anchors.right: Config.options.bar.vertical && Config.options.bar.bottom ? parent.right : (Config.options.bar.vertical ? undefined : parent.right)

                                                anchors.margins: Config.options.bar.cornerStyle === 1 ? 4 : 0

                                                width: Config.options.bar.vertical ? 22 : parent.width - (Config.options.bar.cornerStyle === 1 ? 8 : 0)
                                                height: Config.options.bar.vertical ? parent.height - (Config.options.bar.cornerStyle === 1 ? 8 : 0) : 20

                                                Row {
                                                    anchors.centerIn: parent
                                                    visible: !Config.options.bar.vertical
                                                    spacing: 6
                                                    Text { text: "//"; color: "#ffffff"; font.pixelSize: 8; font.weight: Font.Bold }
                                                    Text { text: "TOREN LINUX"; color: "#ffffff"; font.pixelSize: 8; font.weight: Font.Bold }
                                                }
                                            }

                                            // Mockup Windows
                                            Rectangle {
                                                anchors.centerIn: parent
                                                width: 150
                                                height: 95
                                                radius: 6
                                                color: "#181820"
                                                border.width: 1
                                                border.color: "#30303e"

                                                Row {
                                                    anchors.top: parent.top
                                                    anchors.left: parent.left
                                                    anchors.margins: 6
                                                    spacing: 3
                                                    Rectangle { width: 5; height: 5; radius: 2.5; color: "#e2342a" }
                                                    Rectangle { width: 5; height: 5; radius: 2.5; color: "#e5a50a" }
                                                    Rectangle { width: 5; height: 5; radius: 2.5; color: "#57e389" }
                                                }
                                            }
                                        }
                                    }

                                    StyledText {
                                        Layout.alignment: Qt.AlignHCenter
                                        text: Translation.tr("Desktop Silhouette")
                                        font.pixelSize: Appearance.font.pixelSize.smaller
                                        color: Appearance.colors.colSubtext
                                    }
                                }
                            }
                        }
                    }
                }

                // ------------------------------------------
                // PAGE 2: APPS & DEV STACK
                // ------------------------------------------
                Item {
                    id: pageApps

                    ColumnLayout {
                        anchors.fill: parent
                        spacing: 12

                        ColumnLayout {
                            spacing: 4
                            StyledText {
                                text: Translation.tr("Apps & Dev Stack")
                                font.pixelSize: 24
                                font.weight: Font.Bold
                                color: Appearance.colors.colOnLayer0
                            }
                            StyledText {
                                text: Translation.tr("Toren Linux includes a complete desktop suite and modern developer environment out of the box.")
                                font.pixelSize: Appearance.font.pixelSize.small
                                color: Appearance.colors.colSubtext
                            }
                        }

                        Rectangle {
                            Layout.fillWidth: true
                            height: 1
                            color: Appearance.colors.colOutlineVariant
                        }

                        // Categories Grid (2 Columns)
                        RowLayout {
                            Layout.fillWidth: true
                            Layout.fillHeight: true
                            spacing: 16

                            // Column 1: Core Applications
                            ColumnLayout {
                                Layout.fillWidth: true
                                Layout.fillHeight: true
                                spacing: 8

                                RowLayout {
                                    spacing: 6
                                    MaterialSymbol {
                                        text: "apps"
                                        iconSize: 16
                                        color: Appearance.m3colors.m3primary
                                    }
                                    StyledText {
                                        text: Translation.tr("Core Applications")
                                        font.weight: Font.Bold
                                        font.pixelSize: 13
                                        color: Appearance.colors.colOnLayer0
                                    }
                                }

                                AppToolCard {
                                    iconName: "travel_explore"
                                    titleText: "Chromium"
                                    badgeText: Translation.tr("Browser")
                                    descText: Translation.tr("GPU-accelerated Wayland web browser")
                                    command: ["toren-app", "browser"]
                                }

                                AppToolCard {
                                    iconName: "terminal"
                                    titleText: "Ghostty"
                                    badgeText: Translation.tr("Terminal")
                                    descText: Translation.tr("Fast, GPU-accelerated modern terminal")
                                    command: ["toren-app", "terminal"]
                                }

                                AppToolCard {
                                    iconName: "edit_note"
                                    titleText: "Neovim"
                                    badgeText: Translation.tr("Editor")
                                    descText: Translation.tr("Extensible modal terminal code editor")
                                    command: ["toren-app", "editor"]
                                }

                                AppToolCard {
                                    iconName: "folder_open"
                                    titleText: "Thunar"
                                    badgeText: Translation.tr("Files")
                                    descText: Translation.tr("Lightweight file manager with archive plugins")
                                    command: ["toren-app", "files"]
                                }

                                AppToolCard {
                                    iconName: "shopping_bag"
                                    titleText: "Toren Store"
                                    badgeText: Translation.tr("Store")
                                    descText: Translation.tr("Package manager, components & shell extensions")
                                    command: ["toren-store"]
                                }

                                AppToolCard {
                                    iconName: "settings"
                                    titleText: "Toren Settings"
                                    badgeText: Translation.tr("Control")
                                    descText: Translation.tr("System preferences, displays, audio & keybinds")
                                    command: ["toren-shell", "hub", "open"]
                                }

                                Item { Layout.fillHeight: true }
                            }

                            // Column 2: Developer Tooling & CLI
                            ColumnLayout {
                                Layout.fillWidth: true
                                Layout.fillHeight: true
                                spacing: 8

                                RowLayout {
                                    spacing: 6
                                    MaterialSymbol {
                                        text: "terminal"
                                        iconSize: 16
                                        color: Appearance.m3colors.m3primary
                                    }
                                    StyledText {
                                        text: Translation.tr("Developer Stack & CLI")
                                        font.weight: Font.Bold
                                        font.pixelSize: 13
                                        color: Appearance.colors.colOnLayer0
                                    }
                                }

                                AppToolCard {
                                    iconName: "edit_note"
                                    titleText: "Neovim"
                                    badgeText: Translation.tr("CLI Editor")
                                    descText: Translation.tr("Modal terminal editor with extensible LSP plugins")
                                    command: ["toren-app", "terminal", "--", "nvim"]
                                }

                                AppToolCard {
                                    iconName: "speed"
                                    titleText: "btop"
                                    badgeText: Translation.tr("Monitor")
                                    descText: Translation.tr("Modern resource monitor for CPU, GPU, RAM & nets")
                                    command: ["toren-app", "terminal", "--", "btop"]
                                }

                                AppToolCard {
                                    iconName: "extension"
                                    titleText: "Mise"
                                    badgeText: Translation.tr("Runtimes")
                                    descText: Translation.tr("Polyglot toolchain manager (Node, Python, Go, Rust)")
                                    command: ["toren-app", "terminal", "--", "mise"]
                                }

                                AppToolCard {
                                    iconName: "deployed_code"
                                    titleText: "Docker & Podman"
                                    badgeText: Translation.tr("Containers")
                                    descText: Translation.tr("OCI container engines & Docker Compose included")
                                    command: ["toren-app", "terminal", "--", "podman", "version"]
                                }

                                AppToolCard {
                                    iconName: "terminal"
                                    titleText: "Zsh + Starship"
                                    badgeText: Translation.tr("Shell")
                                    descText: Translation.tr("Fast prompt with syntax highlighting & autosuggest")
                                    command: ["toren-app", "terminal"]
                                }

                                AppToolCard {
                                    iconName: "search"
                                    titleText: "Fastfetch & Atuin"
                                    badgeText: Translation.tr("CLI Tools")
                                    descText: Translation.tr("System specs overview and shell history search")
                                    command: ["toren-app", "terminal", "--", "fastfetch"]
                                }

                                Item { Layout.fillHeight: true }
                            }
                        }

                        NoticeBox {
                            Layout.fillWidth: true
                            text: Translation.tr("Click any card to launch it now. Open the app launcher anytime with <tt>Super+Space</tt>.")
                        }
                    }
                }

                // ------------------------------------------
                // PAGE 3: ESSENTIALS (KEYBINDINGS)
                // ------------------------------------------
                Item {
                    id: pageEssentials

                    ColumnLayout {
                        anchors.fill: parent
                        spacing: 16

                        ColumnLayout {
                            spacing: 4
                            StyledText {
                                text: Translation.tr("The keys that matter")
                                font.pixelSize: 24
                                font.weight: Font.Bold
                                color: Appearance.colors.colOnLayer0
                            }
                            StyledText {
                                text: Translation.tr("These core shortcuts drive everyday tasks. Click any card to launch the tool right now.")
                                font.pixelSize: Appearance.font.pixelSize.small
                                color: Appearance.colors.colSubtext
                            }
                        }

                        Rectangle {
                            Layout.fillWidth: true
                            height: 1
                            color: Appearance.colors.colOutlineVariant
                        }

                        // Shortcuts Grid (2 cols x 4 rows)
                        GridLayout {
                            Layout.fillWidth: true
                            Layout.fillHeight: true
                            columns: 2
                            rowSpacing: 10
                            columnSpacing: 12

                            ShortcutCard {
                                comboText: "Super + Space"
                                titleText: Translation.tr("App Launcher")
                                descText: Translation.tr("Search apps and commands")
                                command: ["toren-shell", "launcher"]
                            }

                            ShortcutCard {
                                comboText: "Super + Enter"
                                titleText: Translation.tr("Terminal")
                                descText: Translation.tr("Ghostty shell window")
                                command: ["toren-app", "terminal"]
                            }

                            ShortcutCard {
                                comboText: "Super + B"
                                titleText: Translation.tr("Browser")
                                descText: Translation.tr("Default Chromium web browser")
                                command: ["toren-app", "browser"]
                            }

                            ShortcutCard {
                                comboText: "Super + I"
                                titleText: Translation.tr("Toren Settings")
                                descText: Translation.tr("System preferences & devices")
                                command: ["toren-shell", "hub", "open"]
                            }

                            ShortcutCard {
                                comboText: "Ctrl+Shift+Esc"
                                titleText: Translation.tr("System Monitor")
                                descText: Translation.tr("btop task manager")
                                command: ["toren-app", "terminal", "--", "btop"]
                            }

                            ShortcutCard {
                                comboText: "Super + Esc"
                                titleText: Translation.tr("Quick Controls")
                                descText: Translation.tr("Audio, network, power menu")
                                command: ["toren-shell", "quicksettings"]
                            }

                            ShortcutCard {
                                comboText: "Super+Shift+S"
                                titleText: Translation.tr("Screen Snip")
                                descText: Translation.tr("Area screenshot to clipboard")
                                command: ["toren-cmd-snip"]
                            }

                            ShortcutCard {
                                comboText: "Super + L"
                                titleText: Translation.tr("Lock Session")
                                descText: Translation.tr("Lock the Wayland display")
                                command: ["toren-shell", "lock"]
                            }
                        }
                    }
                }

                // ------------------------------------------
                // PAGE 4: DONE
                // ------------------------------------------
                Item {
                    id: pageDone

                    Rectangle {
                        anchors.centerIn: parent
                        width: Math.min(parent.width, 680)
                        height: Math.min(parent.height, 420)
                        radius: Appearance.rounding.cardRounding
                        color: Appearance.m3colors.m3surfaceContainerLowest
                        border.width: 1
                        border.color: Appearance.colors.colOutlineVariant

                        ColumnLayout {
                            anchors.centerIn: parent
                            width: Math.min(parent.width - 64, 520)
                            spacing: 20

                            RowLayout {
                                Layout.alignment: Qt.AlignHCenter
                                spacing: 8
                                Text {
                                    text: "//"
                                    color: "#e2342a"
                                    font.family: Appearance.font.family.title
                                    font.pixelSize: 42
                                    font.weight: Font.Black
                                }
                                StyledText {
                                    text: "TOREN LINUX"
                                    font.pixelSize: 32
                                    font.weight: Font.Bold
                                    color: Appearance.colors.colOnLayer0
                                }
                            }

                            StyledText {
                                Layout.fillWidth: true
                                text: Translation.tr("The desktop is yours")
                                font.pixelSize: 26
                                font.weight: Font.Bold
                                horizontalAlignment: Text.AlignHCenter
                                color: Appearance.colors.colOnLayer0
                            }

                            StyledText {
                                Layout.fillWidth: true
                                text: Translation.tr("You are all set! You can reopen this guide anytime with <tt>Super+Shift+Alt+/</tt>, or customize displays, power profiles, and keybinds in Toren Settings (<tt>Super+I</tt>).")
                                font.pixelSize: Appearance.font.pixelSize.medium
                                horizontalAlignment: Text.AlignHCenter
                                wrapMode: Text.WordWrap
                                color: Appearance.colors.colSubtext
                                lineHeight: 1.4
                            }

                            Rectangle {
                                Layout.fillWidth: true
                                height: 1
                                color: Appearance.colors.colOutlineVariant
                            }

                            RowLayout {
                                Layout.alignment: Qt.AlignHCenter
                                spacing: 14

                                RippleButtonWithIcon {
                                    materialIcon: "settings"
                                    mainText: Translation.tr("Open Settings")
                                    onClicked: root.runCmd(["toren-shell", "hub", "open"])
                                }

                                RippleButtonWithIcon {
                                    materialIcon: "terminal"
                                    mainText: Translation.tr("Open Terminal")
                                    onClicked: root.runCmd(["toren-app", "terminal"])
                                }

                                RippleButtonWithIcon {
                                    materialIcon: "check"
                                    mainText: Translation.tr("Finish")
                                    colBackground: Appearance.m3colors.m3primary
                                    onClicked: root.finish()
                                }
                            }
                        }
                    }
                }
            }
        }

        // ==========================================
        // 3. BOTTOM NAVIGATION BAR
        // ==========================================
        Rectangle {
            Layout.fillWidth: true
            implicitHeight: 52
            color: Appearance.m3colors.m3surfaceContainerLow
            radius: Appearance.rounding.windowRounding - 2

            RowLayout {
                anchors.fill: parent
                anchors.leftMargin: 20
                anchors.rightMargin: 20

                // Back / Skip Button
                RippleButtonWithIcon {
                    materialIcon: root.currentStep === 0 ? "close" : "arrow_back"
                    mainText: root.currentStep === 0 ? Translation.tr("Skip Tour") : Translation.tr("Back")
                    onClicked: {
                        if (root.currentStep === 0) {
                            root.finish();
                        } else {
                            root.back();
                        }
                    }
                }

                Item { Layout.fillWidth: true }

                // Step count indicator
                StyledText {
                    text: Translation.tr("Step 0%1 of 0%2").arg(root.currentStep + 1).arg(root.totalSteps)
                    font.pixelSize: Appearance.font.pixelSize.small
                    font.weight: Font.DemiBold
                    color: Appearance.colors.colSubtext
                }

                Item { Layout.fillWidth: true }

                // Next / Finish Button
                RippleButtonWithIcon {
                    materialIcon: root.currentStep === root.totalSteps - 1 ? "check" : "arrow_forward"
                    mainText: root.currentStep === root.totalSteps - 1 ? Translation.tr("Finish") : Translation.tr("Next")
                    colBackground: root.currentStep === root.totalSteps - 1 ? Appearance.m3colors.m3primary : Appearance.m3colors.m3surfaceContainerHigh
                    onClicked: root.advance()
                }
            }
        }
    }
}
