pragma ComponentBehavior: Bound
import qs
import qs.modules.common
import qs.modules.common.functions
import qs.modules.common.widgets
import qs.services
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Qt5Compat.GraphicalEffects
import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import Quickshell.Hyprland

// Options toolbar
Toolbar {
    id: root

    // Use a synchronizer on these
    property var action
    property var selectionMode
    // Signals
    signal dismiss()

    ToolbarTabBar {
        id: actionTabBar
        visible: root.action !== RegionSelection.SnipAction.Record && root.action !== RegionSelection.SnipAction.RecordWithSound
        tabButtonList: [
            {"icon": "content_copy", "name": Translation.tr("Copy")},
            {"icon": "edit", "name": Translation.tr("Edit")},
            {"icon": "image_search", "name": Translation.tr("Lens")},
            {"icon": "document_scanner", "name": Translation.tr("OCR")}
        ]
        currentIndex: {
            switch (root.action) {
                case RegionSelection.SnipAction.Copy: return 0;
                case RegionSelection.SnipAction.Edit: return 1;
                case RegionSelection.SnipAction.Search: return 2;
                case RegionSelection.SnipAction.CharRecognition: return 3;
                default: return 0;
            }
        }
        onCurrentIndexChanged: {
            if (!visible) return;
            const actions = [
                RegionSelection.SnipAction.Copy,
                RegionSelection.SnipAction.Edit,
                RegionSelection.SnipAction.Search,
                RegionSelection.SnipAction.CharRecognition
            ];
            if (currentIndex >= 0 && currentIndex < actions.length) {
                root.action = actions[currentIndex];
            }
        }
    }

    Rectangle {
        visible: actionTabBar.visible
        Layout.fillHeight: true
        Layout.topMargin: 6
        Layout.bottomMargin: 6
        implicitWidth: 1
        color: Appearance.colors.colOutlineVariant
    }

    ToolbarTabBar {
        id: tabBar
        tabButtonList: [
            {"icon": "activity_zone", "name": Translation.tr("Rect")},
            {"icon": "gesture", "name": Translation.tr("Circle")}
        ]
        currentIndex: root.selectionMode === RegionSelection.SelectionMode.RectCorners ? 0 : 1
        onCurrentIndexChanged: {
            root.selectionMode = currentIndex === 0 ? RegionSelection.SelectionMode.RectCorners : RegionSelection.SelectionMode.Circle;
        }
    }
}
