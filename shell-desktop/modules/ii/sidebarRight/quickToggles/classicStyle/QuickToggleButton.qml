import qs.modules.common
import qs.modules.common.widgets
import QtQuick

GroupButton {
    id: button
    property string buttonIcon
    baseWidth: 32
    baseHeight: 28
    clickedWidth: baseWidth + 10
    toggled: false
    buttonRadius: (altAction && toggled) ? Appearance?.rounding.normal : Math.min(baseHeight, baseWidth) / 2
    buttonRadiusPressed: Appearance?.rounding?.small

    contentItem: MaterialSymbol {
        anchors.centerIn: parent
        iconSize: 18
        fill: button.toggled ? 1 : 0
        color: button.toggled ? Appearance.m3colors.m3onPrimary : Appearance.colors.colOnLayer1
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
        text: button.buttonIcon

        Behavior on color {
            animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
        }
    }

}
