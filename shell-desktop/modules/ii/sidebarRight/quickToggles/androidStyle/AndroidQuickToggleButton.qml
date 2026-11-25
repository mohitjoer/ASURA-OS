import QtQuick
import QtQuick.Layouts
import qs.services
import qs.modules.common
import qs.modules.common.models.quickToggles
import qs.modules.common.functions
import qs.modules.common.widgets

GroupButton {
    id: root
    
    // Info to be passed to by repeater
    required property int buttonIndex
    required property var buttonData
    required property bool expandedSize
    required property real baseCellWidth
    required property real baseCellHeight
    required property real cellSpacing
    required property int cellSize

    // Signals
    signal openMenu()

    // Declared in specific toggles
    property QuickToggleModel toggleModel
    property string name: toggleModel?.name ?? ""
    property string statusText: (toggleModel?.hasStatusText) ? (toggleModel?.statusText || (toggled ? Translation.tr("On") : Translation.tr("Off"))) : ""
    property string tooltipText: toggleModel?.tooltipText ?? ""
    property string buttonIcon: toggleModel?.icon ?? "close"
    property bool available: toggleModel?.available ?? true
    toggled: toggleModel?.toggled ?? false
    property var mainAction: toggleModel?.mainAction ?? null
    altAction: toggleModel?.hasMenu ? (() => root.openMenu()) : (toggleModel?.altAction ?? null)

    // Edit mode state
    property bool editMode: false

    // Sizing shenanigans
    baseWidth: root.baseCellWidth * cellSize + cellSpacing * (cellSize - 1)
    baseHeight: root.baseCellHeight
    enableImplicitWidthAnimation: !editMode && root.mouseArea.containsMouse
    enableImplicitHeightAnimation: !editMode && root.mouseArea.containsMouse
    Behavior on baseWidth {
        animation: Appearance.animation.elementMove.numberAnimation.createObject(this)
    }
    Behavior on baseHeight {
        animation: Appearance.animation.elementMove.numberAnimation.createObject(this)
    }
    opacity: 0
    Component.onCompleted: {
        opacity = 1
    }
    Behavior on opacity {
        animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
    }

    enabled: available || editMode
    padding: 4
    horizontalPadding: 8
    verticalPadding: 4

    colBackground: Appearance.colors.colLayer2
    colBackgroundToggled: Appearance.colors.colPrimary
    colBackgroundToggledHover: Appearance.colors.colPrimaryHover
    colBackgroundToggledActive: Appearance.colors.colPrimaryActive
    buttonRadius: Appearance.rounding.normal
    buttonRadiusPressed: Appearance.rounding.small
    property color colText: (root.toggled && enabled) ? Appearance.colors.colOnPrimary : ColorUtils.transparentize(Appearance.colors.colOnLayer2, enabled ? 0 : 0.7)
    property color colIcon: colText

    onClicked: {
        if (root.expandedSize && root.altAction) root.altAction();
        else root.mainAction();
    }

    contentItem: Item {
        anchors.fill: parent

        // Single cell (size 1): cleanly centered icon
        MaterialSymbol {
            visible: !root.expandedSize
            anchors.centerIn: parent
            iconSize: 20
            fill: root.toggled ? 1 : 0
            color: root.colIcon
            text: root.buttonIcon
        }

        // Expanded cell (size > 1): RowLayout with icon, text, and chevron
        RowLayout {
            visible: root.expandedSize
            anchors.fill: parent
            anchors.leftMargin: root.horizontalPadding
            anchors.rightMargin: root.horizontalPadding
            spacing: 6

            // Icon with clickable circle for toggling
            MouseArea {
                id: iconMouseArea
                visible: root.expandedSize
                hoverEnabled: true
                acceptedButtons: !!root.altAction ? Qt.LeftButton : Qt.NoButton
                Layout.alignment: Qt.AlignVCenter
                implicitHeight: 30
                implicitWidth: 30
                cursorShape: !!root.altAction ? Qt.PointingHandCursor : Qt.ArrowCursor

                onClicked: root.mainAction()

                Rectangle {
                    id: iconBackground
                    anchors.fill: parent
                    radius: height / 2
                    color: root.toggled
                        ? ColorUtils.transparentize(Appearance.colors.colOnPrimary, iconMouseArea.containsMouse ? 0.75 : 0.85)
                        : ColorUtils.transparentize(Appearance.colors.colOnLayer2, iconMouseArea.containsMouse ? 0.85 : 0.92)

                    Behavior on color {
                        animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
                    }

                    MaterialSymbol {
                        anchors.centerIn: parent
                        fill: root.toggled ? 1 : 0
                        iconSize: 18
                        color: root.colIcon
                        text: root.buttonIcon
                    }
                }
            }

            // Text column
            Column {
                Layout.alignment: Qt.AlignVCenter
                Layout.fillWidth: true
                spacing: 0

                StyledText {
                    anchors {
                        left: parent.left
                        right: parent.right
                    }
                    font.pixelSize: Appearance.font.pixelSize.smaller
                    font.weight: 600
                    color: root.colText
                    elide: Text.ElideRight
                    text: root.name
                }

                StyledText {
                    visible: root.statusText !== ""
                    anchors {
                        left: parent.left
                        right: parent.right
                    }
                    font {
                        pixelSize: Appearance.font.pixelSize.smallest
                        weight: 400
                    }
                    color: root.colText
                    opacity: 0.8
                    elide: Text.ElideRight
                    text: root.statusText
                }
            }

            // Menu chevron
            MaterialSymbol {
                visible: !!root.altAction
                Layout.alignment: Qt.AlignVCenter
                Layout.rightMargin: 2
                iconSize: 16
                color: root.colText
                opacity: 0.6
                text: "chevron_right"
            }
        }
    }

    MouseArea { // Blocking MouseArea for edit interactions
        id: editModeInteraction
        visible: root.editMode
        anchors.fill: parent
        cursorShape: Qt.PointingHandCursor
        hoverEnabled: true
        acceptedButtons: Qt.AllButtons

        function toggleEnabled() {
            const index = root.buttonIndex;
            const toggleList = Config.options.sidebar.quickToggles.android.toggles;
            const buttonType = root.buttonData.type;
            if (!toggleList.find(toggle => toggle.type === buttonType)) {
                toggleList.push({ type: buttonType, size: 1 });
            } else {
                toggleList.splice(index, 1);
            }
        }

        function toggleSize() {
            const index = root.buttonIndex;
            const toggleList = Config.options.sidebar.quickToggles.android.toggles;
            const buttonType = root.buttonData.type;
            if (!toggleList.find(toggle => toggle.type === buttonType)) return;
            toggleList[index].size = 3 - toggleList[index].size; // Alternate between 1 and 2
        }

        function movePositionBy(offset) {
            const index = root.buttonIndex;
            const toggleList = Config.options.sidebar.quickToggles.android.toggles;
            const buttonType = root.buttonData.type;
            const targetIndex = index + offset;
            if (!toggleList.find(toggle => toggle.type === buttonType)) return;
            if (targetIndex < 0 || targetIndex >= toggleList.length) return;
            const temp = toggleList[index];
            toggleList[index] = toggleList[targetIndex];
            toggleList[targetIndex] = temp;
        }

        onReleased: (event) => {
            if (event.button === Qt.LeftButton)
                toggleEnabled();
        }
        onPressed: (event) => {
            if (event.button === Qt.RightButton) toggleSize();
        }
        onPressAndHold: (event) => { // Also toggle size
            toggleSize();
        }
        onWheel: (event) => {
            const index = root.buttonIndex;
            const toggleList = Config.options.sidebar.quickToggles.android.toggles;
            const buttonType = root.buttonData.type;
            if (event.angleDelta.y < 0) { // Move to right
                movePositionBy(1);
            } else if (event.angleDelta.y > 0) { // Move to left
                movePositionBy(-1);
            }
            event.accepted = true;
        }
    }

    StyledToolTip {
        extraVisibleCondition: root.tooltipText !== ""
        text: root.tooltipText
    }
}
