import QtQuick
import qs.Ui.Singletons

Item {
    id: root
    property real size: 28
    property string variant: Tokens.light ? "dark" : "white" // "white" (default for dark bg) or "dark" (for light bg)
    implicitWidth: size
    implicitHeight: size

    Image {
        id: img
        anchors.fill: parent
        source: {
            if (root.variant === "dark") {
                return Qt.resolvedUrl("../assets/asura-logo.svg");
            }
            return Qt.resolvedUrl("../assets/asura-logo-on-dark.svg");
        }
        sourceSize.width: Math.round(root.size * 2)
        sourceSize.height: Math.round(root.size * 2)
        fillMode: Image.PreserveAspectFit
        smooth: true
        mipmap: true
    }
}
