import qs.modules.common
import qs.modules.common.widgets
import qs.services
import Qt5Compat.GraphicalEffects
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Item {
    id: root

    NotificationListView { // Scrollable window
        id: listview
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        anchors.bottom: statusRow.top
        anchors.bottomMargin: 5

        clip: true
        layer.enabled: true
        layer.effect: OpacityMask {
            maskSource: Rectangle {
                width: listview.width
                height: listview.height
                radius: Appearance.rounding.normal
            }
        }

        popup: false
    }

    // Placeholder when list is empty
    Item {
        anchors.fill: parent
        anchors.bottomMargin: statusRow.height + 10
        visible: Notifications.list.length === 0
        opacity: visible ? 1 : 0
        Behavior on opacity {
            animation: Appearance.animation.elementMoveEnter.numberAnimation.createObject(this)
        }

        ColumnLayout {
            anchors.centerIn: parent
            spacing: 6

            MaterialSymbol {
                Layout.alignment: Qt.AlignHCenter
                iconSize: 32
                color: Appearance.m3colors.m3outline
                text: "notifications_none"
            }

            StyledText {
                Layout.alignment: Qt.AlignHCenter
                font.pixelSize: Appearance.font.pixelSize.small
                color: Appearance.m3colors.m3outline
                text: Translation.tr("No notifications")
            }
        }
    }

    ButtonGroup {
        id: statusRow
        anchors {
            left: parent.left
            right: parent.right
            bottom: parent.bottom
        }

        NotificationStatusButton {
            Layout.fillWidth: false
            buttonIcon: "notifications_paused"
            toggled: Notifications.silent
            onClicked: () => {
                Notifications.silent = !Notifications.silent;
            }
        }
        NotificationStatusButton {
            enabled: false
            Layout.fillWidth: true
            buttonText: Translation.tr("%1 notifications").arg(Notifications.list.length)
        }
        NotificationStatusButton {
            Layout.fillWidth: false
            buttonIcon: "delete_sweep"
            onClicked: () => {
                Notifications.discardAllNotifications()
            }
        }
    }
}