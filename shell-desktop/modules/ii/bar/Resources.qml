import qs.modules.common
import qs.services
import QtQuick
import QtQuick.Layouts

// ASURA overlay (shell-desktop/asura/vendor.sh OVERLAYS): the bar's resources,
// at the right end of the bar. CPU and RAM are always shown, swap only while
// it is in use, each as a labelled readout (Resource.qml). Hovering shows
// end-4's detailed popup.
MouseArea {
    id: root
    property bool borderless: Config.options.bar.borderless
    property bool alwaysShowAllResources: false
    implicitWidth: rowLayout.implicitWidth + 12 * 2
    implicitHeight: Appearance.sizes.barHeight
    hoverEnabled: !Config.options.bar.tooltips.clickToShow

    RowLayout {
        id: rowLayout
        anchors.centerIn: parent
        spacing: 14

        Resource {
            iconName: "planner_review"
            label: "CPU"
            percentage: ResourceUsage.cpuUsage
            warningThreshold: Config.options.bar.resources.cpuWarningThreshold
        }
        Resource {
            iconName: "memory"
            label: "RAM"
            percentage: ResourceUsage.memoryUsedPercentage
            warningThreshold: Config.options.bar.resources.memoryWarningThreshold
        }
        Resource {
            iconName: "swap_horiz"
            label: "SWAP"
            percentage: ResourceUsage.swapUsedPercentage
            shown: percentage > 0.005 || root.alwaysShowAllResources
            warningThreshold: Config.options.bar.resources.swapWarningThreshold
        }
    }

    ResourcesPopup {
        hoverTarget: root
    }
}
