import QtQuick
import "Singletons"
import qs.Ui.Singletons

// The bar every settings page needs before it needs anything else. A page that
// previews live and cannot save does not look broken, it looks fine and then
// eats the edit on the way out, so this exists as one component rather than as
// something each page remembers to build.
//
// Revert and Reset are not the same verb and the old pages conflated them:
//   revert  discards unsaved edits, back to what is on disk
//   reset   sets every key to its factory value, which is itself an edit
// Reset is destructive and sits behind a divider, away from Save.
Item {
    id: bar

    property int dirty: 0
    property string cleanText: I18n.tr("SAVED · LIVE ON YOUR DESKTOP")
    property string savingText: ""

    signal saved()
    signal reverted()
    signal reset()
    signal diffRequested()

    implicitHeight: Tokens.cards ? 68 : 60

    // the card look writes the status in sentence case ("Saved · live on ...")
    function sentence(t) {
        t = String(t || "");
        if (!Tokens.cards || t !== t.toUpperCase()) return t;
        return t.charAt(0) + t.slice(1).toLowerCase();
    }

    Rectangle { visible: !Tokens.cards; anchors.top: parent.top; width: parent.width; height: 1; color: Tokens.line }
    // the card look: the bar is a soft rounded card floating above the bottom
    // edge, like the setting cards above it
    Rectangle {
        id: plate
        visible: Tokens.cards
        anchors.fill: parent
        anchors.leftMargin: Tokens.s5; anchors.rightMargin: Tokens.s5
        anchors.topMargin: Tokens.s2; anchors.bottomMargin: Tokens.s3
        radius: Tokens.radiusCard
        color: Tokens.tint5
    }

    Row {
        anchors.verticalCenter: parent.verticalCenter
        anchors.verticalCenterOffset: Tokens.cards ? Math.round((Tokens.s2 - Tokens.s3) / 2) : 0
        anchors.left: parent.left
        anchors.leftMargin: Tokens.cards ? Tokens.s5 + Tokens.s4 : Tokens.s5
        spacing: Tokens.s3

        Text {
            visible: !Tokens.cards
            anchors.verticalCenter: parent.verticalCenter
            text: "///"
            color: Tokens.inkFaint
            font.family: Tokens.mono
            font.pixelSize: 10
        }
        Rectangle {
            width: Tokens.cards ? 8 : 6; height: width
            radius: Tokens.cards ? width / 2 : 0
            anchors.verticalCenter: parent.verticalCenter
            // no colour in the chrome: the unsaved heartbeat is ink, and the
            // /// cluster beside it carries the sheet's mark. The card look
            // lights it with the accent while there are unsaved changes.
            color: Tokens.cards ? (bar.dirty > 0 ? Tokens.sun : Tokens.inkFaint) : Tokens.ink
            SequentialAnimation on opacity {
                running: bar.dirty > 0
                loops: Animation.Infinite
                NumberAnimation { to: 0.3; duration: 600 }
                NumberAnimation { to: 1.0; duration: 600 }
            }
        }
        Text {
            anchors.verticalCenter: parent.verticalCenter
            text: bar.sentence(bar.dirty > 0
                  ? (bar.dirty === 1 ? I18n.tr("%1 CHANGE · PREVIEWING · NOT SAVED") : I18n.tr("%1 CHANGES · PREVIEWING · NOT SAVED")).arg(bar.dirty)
                  : I18n.tr(bar.cleanText))
            color: bar.dirty > 0 ? Tokens.ink : Tokens.inkDim
            font.family: Tokens.ui
            font.pixelSize: Tokens.cards ? 12 : 11
            font.weight: Font.Medium
            font.letterSpacing: Tokens.cards ? 0 : 1.6
        }
        Text {
            visible: bar.dirty > 0
            anchors.verticalCenter: parent.verticalCenter
            text: Tokens.cards ? I18n.tr("View changes")
                : (sdh.hovered ? I18n.tr("\u25b8 VIEW DIFF") : I18n.tr("\u00b7 VIEW DIFF"))
            color: Tokens.cards ? (sdh.hovered ? Tokens.ink : Tokens.sun) : (sdh.hovered ? Tokens.ink : Tokens.inkFaint)
            font.family: Tokens.mono; font.pixelSize: Tokens.cards ? 12 : 9
            font.underline: Tokens.cards && sdh.hovered
            font.letterSpacing: Tokens.cards ? 0 : 1.2
        }
        HoverHandler { id: sdh; enabled: bar.dirty > 0; cursorShape: Qt.PointingHandCursor }
        TapHandler { enabled: bar.dirty > 0; onTapped: bar.diffRequested() }
    }

    Row {
        anchors.verticalCenter: parent.verticalCenter
        anchors.verticalCenterOffset: Tokens.cards ? Math.round((Tokens.s2 - Tokens.s3) / 2) : 0
        anchors.right: parent.right
        anchors.rightMargin: Tokens.cards ? Tokens.s5 + Tokens.s2 : Tokens.s6
        spacing: Tokens.cards ? Tokens.s2 : Tokens.s3

        Btn { text: I18n.tr("RESET TO DEFAULTS"); onAct: bar.reset() }
        Rectangle {
            visible: !Tokens.cards
            width: 1; height: 22
            color: Tokens.line
            anchors.verticalCenter: parent.verticalCenter
        }
        Btn { text: I18n.tr("REVERT"); armed: bar.dirty > 0; onAct: bar.reverted() }
        Btn { text: I18n.tr("SAVE"); primary: true; armed: bar.dirty > 0; onAct: bar.saved() }
    }
}
