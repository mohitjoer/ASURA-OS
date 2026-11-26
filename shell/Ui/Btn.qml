import QtQuick
import "Singletons"

Rectangle {
    id: btn
    property string text: ""
    property bool primary: false
    property bool armed: true
    property bool compact: false
    signal act()

    // the card look writes an all-caps label in sentence case ("Save", not "SAVE")
    function sentence(t) {
        t = String(t || "");
        if (!Tokens.cards || t !== t.toUpperCase() || t.toLowerCase() === t) return t;
        return t.charAt(0) + t.slice(1).toLowerCase();
    }

    activeFocusOnTab: armed
    function activate() { if (armed) act() }
    Keys.onPressed: event => {
        if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter || event.key === Qt.Key_Space) {
            activate();
            event.accepted = true;
        }
    }

    implicitWidth: lab.width + (compact ? 20 : 30)
    implicitHeight: compact ? 24 : (Tokens.cards ? 34 : 32)
    // the card look: a rounded pill, the primary one filled with the accent
    radius: Tokens.cards ? height / 2 : Tokens.radius
    // 0.5, not 0.3: a control that cannot act must still be legible as itself
    // (the reader needs to see what is there and why it is inert), while the
    // dimming keeps it from reading as available.
    opacity: armed ? 1 : 0.5
    color: Tokens.cards
        ? (primary && armed ? Tokens.sun : (tap.pressed && armed ? Tokens.tint16 : (bh.hovered && armed ? Tokens.tint16 : Tokens.tint10)))
        : (primary && armed ? Tokens.bone : (tap.pressed && armed ? Tokens.tint16 : (bh.hovered && armed ? Tokens.tint10 : "transparent")))
    border.width: Tokens.cards ? (activeFocus ? 1 : 0) : Tokens.border
    border.color: activeFocus ? Tokens.bone : (primary && armed ? Tokens.bone : (bh.hovered && armed ? Tokens.lineStrong : Tokens.line))
    Behavior on color { ColorAnimation { duration: Tokens.snap } }
    Behavior on opacity { NumberAnimation { duration: Tokens.snap } }

    Text {
        id: lab
        anchors.centerIn: parent
        text: btn.sentence(btn.text)
        color: btn.primary && btn.armed ? (Tokens.cards ? Tokens.onSun : Tokens.inkOnBone) : Tokens.ink
        font.family: Tokens.ui
        font.pixelSize: btn.compact ? 10 : (Tokens.cards ? 12 : 11)
        font.weight: Font.Medium
        font.letterSpacing: Tokens.cards ? 0 : Tokens.trackLabel
    }
    HoverHandler { id: bh; enabled: btn.armed; cursorShape: Qt.PointingHandCursor }
    TapHandler { id: tap; enabled: btn.armed; onTapped: btn.act() }
}
