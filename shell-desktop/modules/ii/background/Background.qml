pragma ComponentBehavior: Bound

import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.widgets.widgetCanvas
import qs.modules.common.functions as CF
import QtQuick
import QtQuick.Layouts
import Qt5Compat.GraphicalEffects
import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import Quickshell.Hyprland

import qs.modules.ii.background.widgets
import qs.modules.ii.background.widgets.weather

Variants {
    id: root
    model: Quickshell.screens

    PanelWindow {
        id: bgRoot

        required property var modelData

        // Hide when fullscreen
        property list<HyprlandWorkspace> workspacesForMonitor: Hyprland.workspaces.values.filter(workspace => workspace.monitor && workspace.monitor.name == monitor.name)
        property var activeWorkspaceWithFullscreen: workspacesForMonitor.filter(workspace => ((workspace.toplevels.values.filter(window => window.wayland?.fullscreen)[0] != undefined) && workspace.active))[0]
        visible: GlobalStates.screenLocked || (!(activeWorkspaceWithFullscreen != undefined)) || !Config?.options.background.hideWhenFullscreen

        // Workspaces
        property HyprlandMonitor monitor: Hyprland.monitorFor(modelData)
        property list<var> relevantWindows: HyprlandData.windowList.filter(win => win.monitor == monitor?.id && win.workspace.id >= 0).sort((a, b) => a.workspace.id - b.workspace.id)
        property int firstWorkspaceId: relevantWindows[0]?.workspace.id || 1
        property int lastWorkspaceId: relevantWindows[relevantWindows.length - 1]?.workspace.id || 10
        property int workspaceChunkSize: Config?.options.bar.workspaces.shown ?? 10
        property int totalWorkspaces: Math.ceil(lastWorkspaceId / workspaceChunkSize) * workspaceChunkSize
        // Wallpaper
        property bool wallpaperIsVideo: Config.options.background.wallpaperPath.endsWith(".mp4") || Config.options.background.wallpaperPath.endsWith(".webm") || Config.options.background.wallpaperPath.endsWith(".mkv") || Config.options.background.wallpaperPath.endsWith(".avi") || Config.options.background.wallpaperPath.endsWith(".mov")
        property string wallpaperPath: wallpaperIsVideo ? Config.options.background.thumbnailPath : Config.options.background.wallpaperPath
        property bool wallpaperSafetyTriggered: {
            const enabled = Config.options.workSafety.enable.wallpaper;
            const sensitiveWallpaper = (CF.StringUtils.stringListContainsSubstring(wallpaperPath.toLowerCase(), Config.options.workSafety.triggerCondition.fileKeywords));
            const sensitiveNetwork = (CF.StringUtils.stringListContainsSubstring(Network.networkName.toLowerCase(), Config.options.workSafety.triggerCondition.networkNameKeywords));
            return enabled && sensitiveWallpaper && sensitiveNetwork;
        }
        readonly property bool parallaxActive: Config.options.background.parallax.enableWorkspace || Config.options.background.parallax.enableSidebar
        readonly property real parallaxRation: parallaxActive ? Config.options.background.parallax.workspaceZoom : 1.0
        property real minSuitableScale: 1 // Some reasonable init, to be updated
        property real effectiveWallpaperScale: minSuitableScale * parallaxRation
        property int wallpaperWidth: modelData.width // Some reasonable init value, to be updated
        property int wallpaperHeight: modelData.height // Some reasonable init value, to be updated
        property real scaledWallpaperWidth: wallpaperWidth * effectiveWallpaperScale
        property real scaledWallpaperHeight: wallpaperHeight * effectiveWallpaperScale
        property real parallaxTotalPixelsX: Math.max(0, scaledWallpaperWidth - screen.width)
        property real parallaxTotalPixelsY: Math.max(0, scaledWallpaperHeight - screen.height)
        readonly property bool verticalParallax: (Config.options.background.parallax.autoVertical && wallpaperHeight > wallpaperWidth) || Config.options.background.parallax.vertical
        // Colors
        property bool shouldBlur: (GlobalStates.screenLocked && Config.options.lock.blur.enable)
        property color dominantColor: (bgQuantizer.colors && bgQuantizer.colors.length > 0) ? bgQuantizer.colors[0] : (torenColorsAdapter.primary ?? Appearance.colors.colPrimary)
        property bool dominantColorIsDark: dominantColor.hslLightness < 0.55
        property color colText: {
            if (wallpaperSafetyTriggered)
                return CF.ColorUtils.mix(Appearance.colors.colOnLayer0, dominantColor, 0.75);
            return (GlobalStates.screenLocked && shouldBlur) ? Appearance.colors.colOnLayer0 : CF.ColorUtils.colorWithLightness(dominantColor, (dominantColorIsDark ? 0.90 : 0.12));
        }
        Behavior on colText {
            animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
        }

        // Layer props
        screen: modelData
        exclusionMode: ExclusionMode.Ignore
        WlrLayershell.layer: (GlobalStates.screenLocked && !scaleAnim.running) ? WlrLayer.Overlay : WlrLayer.Bottom
        // WlrLayershell.layer: WlrLayer.Bottom
        WlrLayershell.namespace: "quickshell:background"
        anchors {
            top: true
            bottom: true
            left: true
            right: true
        }
        color: {
            if (!bgRoot.wallpaperSafetyTriggered || bgRoot.wallpaperIsVideo)
                return "transparent";
            return CF.ColorUtils.mix(Appearance.colors.colLayer0, Appearance.colors.colPrimary, 0.75);
        }
        Behavior on color {
            animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
        }

        // Wallpaper color quantizer: samples dominant color directly from active wallpaper
        ColorQuantizer {
            id: bgQuantizer
            source: Qt.resolvedUrl(bgRoot.wallpaperPath)
            depth: 0 // 1 dominant color
            rescaleSize: 16
        }

        // Cache colors file reader
        FileView {
            id: torenColorsFile
            path: (Quickshell.env("XDG_CACHE_HOME") || (Quickshell.env("HOME") + "/.cache")) + "/toren/colors.json"
            watchChanges: true
            printErrors: false
            onFileChanged: reload()

            JsonAdapter {
                id: torenColorsAdapter
                property color primary: Appearance.colors.colPrimary
                property color secondary: Appearance.colors.colSecondary
            }
        }

        // ASURA widgets persistence (reads widgets.json configured from Hub Settings > Widgets)
        FileView {
            id: torenWidgetsFile
            path: (Quickshell.env("XDG_CONFIG_HOME") || (Quickshell.env("HOME") + "/.config")) + "/toren/widgets.json"
            watchChanges: true
            printErrors: false
            onFileChanged: reload()

            JsonAdapter {
                id: torenWidgetsAdapter
                property bool clockEnabled: true
                property string clockDesign: "zen"
                property bool clock24h: true
                property bool clockSeconds: false
                property string clockAccent: "palette"
                property real clockScale: 1.0
                property string clockAnchor: "top-left"
                property int clockX: 72
                property int clockY: 64
                property bool clockLocked: false
                property bool dateShow: true
                property string dateDesign: "hud"
            }
        }

        onWallpaperPathChanged: {
            bgRoot.updateZoomScale();
            // Clock position gets updated after zoom scale is updated
        }

        // Wallpaper zoom scale
        function updateZoomScale() {
            getWallpaperSizeProc.path = bgRoot.wallpaperPath;
            getWallpaperSizeProc.running = true;
        }
        Process {
            id: getWallpaperSizeProc
            property string path: bgRoot.wallpaperPath
            command: ["magick", "identify", "-format", "%w %h", path]
            stdout: StdioCollector {
                id: wallpaperSizeOutputCollector
                onStreamFinished: {
                    const output = wallpaperSizeOutputCollector.text;
                    const [width, height] = output.split(" ").map(Number);
                    const [screenWidth, screenHeight] = [bgRoot.screen.width, bgRoot.screen.height];
                    bgRoot.wallpaperWidth = width;
                    bgRoot.wallpaperHeight = height;

                    // Perfect image; scale = 1
                    // Small picture; scale > 1; will zoom in the picture
                    // Big picture; scale < 1; will zoom out the picture
                    // Choose max number so every side will fit
                    bgRoot.minSuitableScale = Math.max(screenWidth / width, screenHeight / height);
                }
            }
        }

        Item {
            anchors.fill: parent

            // Wallpaper
            StyledImage {
                id: wallpaper
                visible: opacity > 0 && !blurLoader.active
                opacity: (status === Image.Ready && !bgRoot.wallpaperIsVideo) ? 1 : 0
                cache: false
                smooth: false

                property int workspaceIndex: (bgRoot.monitor.activeWorkspace?.id ?? 1) - 1
                property real middleFraction: 0.5
                property real fraction: {
                    // 0 - start of the picture
                    // 1 - end of the picture
                    if (bgRoot.totalWorkspaces <= 1) {
                        return middleFraction;
                    }
                    return Math.max(0, Math.min(1, workspaceIndex / (bgRoot.totalWorkspaces - 1)));
                }

                property real usedFractionX: {
                    let usedFraction = middleFraction;
                    if (Config.options.background.parallax.enableWorkspace && !bgRoot.verticalParallax) {
                        usedFraction = fraction;
                    }
                    if (Config.options.background.parallax.enableSidebar) {
                        let sidebarFraction = bgRoot.parallaxRation / bgRoot.workspaceChunkSize / 2;
                        usedFraction += (sidebarFraction * GlobalStates.sidebarRightOpen - sidebarFraction * GlobalStates.sidebarLeftOpen);
                    }
                    return Math.max(0, Math.min(1, usedFraction));
                }
                property real usedFractionY: {
                    let usedFraction = middleFraction;
                    if (Config.options.background.parallax.enableWorkspace && bgRoot.verticalParallax) {
                        usedFraction = fraction;
                    }
                    return Math.max(0, Math.min(1, usedFraction));
                }

                x: {
                    if (!bgRoot.parallaxActive) {
                        return 0;
                    }
                    if (bgRoot.screen.width > width) {
                        // Center the picture
                        return (bgRoot.screen.width - width) / 2;
                    }
                    return - bgRoot.parallaxTotalPixelsX * usedFractionX;
                }
                y: {
                    if (!bgRoot.parallaxActive) {
                        return 0;
                    }
                    if (bgRoot.screen.height > height) {
                        // Center the picture
                        return (bgRoot.screen.height - height) / 2;
                    }
                    return - bgRoot.parallaxTotalPixelsY * usedFractionY;
                }

                source: bgRoot.wallpaperSafetyTriggered ? "" : bgRoot.wallpaperPath
                fillMode: bgRoot.parallaxActive ? Image.PreserveAspectCrop : Image.Stretch
                Behavior on x {
                    NumberAnimation {
                        duration: 600
                        easing.type: Easing.OutCubic
                    }
                }
                Behavior on y {
                    NumberAnimation {
                        duration: 600
                        easing.type: Easing.OutCubic
                    }
                }
                width: bgRoot.parallaxActive ? bgRoot.scaledWallpaperWidth : bgRoot.screen.width
                height: bgRoot.parallaxActive ? bgRoot.scaledWallpaperHeight : bgRoot.screen.height
            }

            Loader {
                id: blurLoader
                active: Config.options.lock.blur.enable && (GlobalStates.screenLocked || scaleAnim.running)
                anchors.fill: wallpaper
                scale: GlobalStates.screenLocked ? Config.options.lock.blur.extraZoom : 1
                Behavior on scale {
                    NumberAnimation {
                        id: scaleAnim
                        duration: 400
                        easing.type: Easing.BezierSpline
                        easing.bezierCurve: Appearance.animationCurves.expressiveDefaultSpatial
                    }
                }
                sourceComponent: GaussianBlur {
                    source: wallpaper
                    radius: GlobalStates.screenLocked ? Config.options.lock.blur.radius : 0
                    samples: radius * 2 + 1

                    Rectangle {
                        opacity: GlobalStates.screenLocked ? 1 : 0
                        anchors.fill: parent
                        color: CF.ColorUtils.transparentize(Appearance.colors.colLayer0, 0.7)
                    }
                }
            }

            WidgetCanvas {
                id: widgetCanvas
                width: parent.width
                height: parent.height
                readonly property real parallaxFactor: {
                    var f = Config.options.background.parallax.widgetsFactor;
                    return f / bgRoot.parallaxRation;
                }
                readonly property real baseWallpaperOffsetX: (bgRoot.screen.width - wallpaper.width) / 2
                readonly property real baseWallpaperOffsetY: (bgRoot.screen.height - wallpaper.height) / 2
                readonly property real wallpaperTotalOffsetX: wallpaper.x - baseWallpaperOffsetX
                readonly property real wallpaperTotalOffsetY: wallpaper.y - baseWallpaperOffsetY
                readonly property bool locked: GlobalStates.screenLocked
                x: wallpaperTotalOffsetX * parallaxFactor * !locked
                y: wallpaperTotalOffsetY * parallaxFactor * !locked

                transitions: Transition {
                    PropertyAnimation {
                        properties: "width,height"
                        duration: Appearance.animation.elementMove.duration
                        easing.type: Appearance.animation.elementMove.type
                        easing.bezierCurve: Appearance.animation.elementMove.bezierCurve
                    }
                    AnchorAnimation {
                        duration: Appearance.animation.elementMove.duration
                        easing.type: Appearance.animation.elementMove.type
                        easing.bezierCurve: Appearance.animation.elementMove.bezierCurve
                    }
                }

                FadeLoader {
                    shown: Config.options.background.widgets.weather.enable
                    sourceComponent: WeatherWidget {
                        screenWidth: bgRoot.screen.width
                        screenHeight: bgRoot.screen.height
                        scaledScreenWidth: bgRoot.screen.width
                        scaledScreenHeight: bgRoot.screen.height
                        wallpaperScale: 1
                    }
                }

                // ASURA Desktop HUD Clock: Default Wallpaper Widget
                FadeLoader {
                    id: torenClockLoader
                    shown: torenWidgetsAdapter.clockEnabled && (!GlobalStates.screenLocked || Config.options.lock.centerClock === false)
                    sourceComponent: Item {
                        id: clockWidget
                        width: clockCol.implicitWidth
                        height: clockCol.implicitHeight

                        readonly property real s: Math.max(0.5, Math.min(2.0, torenWidgetsAdapter.clockScale || 1.0))
                        readonly property string anchor: torenWidgetsAdapter.clockAnchor || "top-left"
                        readonly property real marginX: 72
                        readonly property real marginY: 64

                        x: {
                            if (anchor === "free") return Math.max(0, Math.min(torenWidgetsAdapter.clockX, bgRoot.screen.width - width));
                            if (anchor.indexOf("right") >= 0) return bgRoot.screen.width - width - marginX;
                            if (anchor.indexOf("center") >= 0 || anchor === "top" || anchor === "bottom") return (bgRoot.screen.width - width) / 2;
                            return marginX;
                        }
                        y: {
                            if (anchor === "free") return Math.max(0, Math.min(torenWidgetsAdapter.clockY, bgRoot.screen.height - height));
                            if (anchor.indexOf("bottom") >= 0) return bgRoot.screen.height - height - marginY;
                            if (anchor === "center" || anchor === "left" || anchor === "right") return (bgRoot.screen.height - height) / 2;
                            return marginY;
                        }

                        property var now: new Date()
                        Timer {
                            interval: 1000
                            running: true
                            repeat: true
                            triggeredOnStart: true
                            onTriggered: clockWidget.now = new Date()
                        }

                        function pad2(n) { return (n < 10 ? "0" : "") + n; }
                        readonly property int rawH: now.getHours()
                        readonly property int mins: now.getMinutes()
                        readonly property int secs: now.getSeconds()
                        readonly property int h12: (rawH % 12) === 0 ? 12 : (rawH % 12)
                        readonly property string hh: torenWidgetsAdapter.clock24h ? pad2(rawH) : String(h12)
                        readonly property string mm: pad2(mins)
                        readonly property string ss: pad2(secs)
                        readonly property string ampm: rawH < 12 ? "AM" : "PM"

                        readonly property var weekdaysShort: ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"]
                        readonly property string dateStr: weekdaysShort[now.getDay()] + " " + now.getDate()

                        readonly property color wallAccent: {
                            var p = (torenColorsAdapter.primary && torenColorsAdapter.primary != "#000000") ? torenColorsAdapter.primary : bgRoot.dominantColor;
                            var hue = p.hsvHue < 0 ? bgRoot.dominantColor.hsvHue : p.hsvHue;
                            if (hue < 0) hue = 0.58; // clean cyan-blue fallback
                            var sat = Math.max(0.45, Math.min(1.0, p.hsvSaturation * 1.35));
                            var val = bgRoot.dominantColorIsDark ? 0.92 : 0.20;
                            return Qt.hsva(hue, sat, val, 1.0);
                        }
                        readonly property color ink: bgRoot.colText || (bgRoot.dominantColorIsDark ? "#FFFFFF" : "#1A1A1A")
                        readonly property color inkSoft: CF.ColorUtils.transparentize(ink, 0.20)
                        readonly property color inkDim: CF.ColorUtils.transparentize(ink, 0.45)
                        readonly property color accent: {
                            if (torenWidgetsAdapter.clockAccent === "brand") return "#F25623";
                            if (torenWidgetsAdapter.clockAccent === "mono") return ink;
                            return wallAccent;
                        }

                        Column {
                            id: clockCol
                            spacing: Math.round(14 * clockWidget.s)

                            // Zen Clock Face: 15 | 59
                            Row {
                                id: faceRow
                                anchors.horizontalCenter: parent.horizontalCenter
                                spacing: Math.round(10 * clockWidget.s)

                                Text {
                                    text: clockWidget.hh
                                    color: clockWidget.ink
                                    font.family: Appearance.font.family.main
                                    font.pixelSize: Math.round(56 * clockWidget.s)
                                    font.weight: Font.Light
                                    font.letterSpacing: -1
                                    style: Text.Raised
                                    styleColor: Qt.rgba(0, 0, 0, 0.45)
                                }

                                Rectangle {
                                    anchors.verticalCenter: parent.verticalCenter
                                    width: Math.max(1, Math.round(2 * clockWidget.s))
                                    height: Math.round(38 * clockWidget.s)
                                    radius: 1
                                    color: clockWidget.accent
                                    opacity: 0.85
                                }

                                Text {
                                    text: clockWidget.mm
                                    color: clockWidget.accent
                                    font.family: Appearance.font.family.main
                                    font.pixelSize: Math.round(56 * clockWidget.s)
                                    font.weight: Font.DemiBold
                                    font.letterSpacing: -1
                                    style: Text.Raised
                                    styleColor: Qt.rgba(0, 0, 0, 0.45)
                                }

                                Column {
                                    visible: torenWidgetsAdapter.clockSeconds || !torenWidgetsAdapter.clock24h
                                    anchors.verticalCenter: parent.verticalCenter
                                    spacing: 2
                                    Text {
                                        visible: torenWidgetsAdapter.clockSeconds
                                        text: clockWidget.ss
                                        color: clockWidget.accent
                                        font.family: Appearance.font.family.mono
                                        font.pixelSize: Math.round(16 * clockWidget.s)
                                        font.weight: Font.Medium
                                    }
                                    Text {
                                        visible: !torenWidgetsAdapter.clock24h
                                        text: clockWidget.ampm
                                        color: clockWidget.inkDim
                                        font.family: Appearance.font.family.main
                                        font.pixelSize: Math.round(13 * clockWidget.s)
                                        font.weight: Font.DemiBold
                                    }
                                }
                            }

                            // HUD Date & Vitals Pill: [ FRI 20 | 94% | 12% CPU · 4.2G ]
                            Rectangle {
                                id: hudPill
                                visible: torenWidgetsAdapter.dateShow
                                anchors.horizontalCenter: parent.horizontalCenter
                                implicitWidth: hudRow.implicitWidth + Math.round(22 * clockWidget.s)
                                implicitHeight: hudRow.implicitHeight + Math.round(10 * clockWidget.s)
                                radius: Math.round(8 * clockWidget.s)
                                color: bgRoot.dominantColorIsDark ? Qt.rgba(0.04, 0.06, 0.10, 0.55) : Qt.rgba(1, 1, 1, 0.65)
                                border.width: 1
                                border.color: Qt.rgba(clockWidget.accent.r, clockWidget.accent.g, clockWidget.accent.b, 0.35)

                                Row {
                                    id: hudRow
                                    anchors.centerIn: parent
                                    spacing: Math.round(8 * clockWidget.s)

                                    // Date
                                    Text {
                                        anchors.verticalCenter: parent.verticalCenter
                                        text: clockWidget.dateStr
                                        color: clockWidget.accent
                                        font.family: Appearance.font.family.main
                                        font.pixelSize: Math.round(12 * clockWidget.s)
                                        font.weight: Font.DemiBold
                                        style: Text.Raised
                                        styleColor: Qt.rgba(0, 0, 0, 0.45)
                                    }

                                    // Divider
                                    Rectangle {
                                        width: 1
                                        height: Math.round(12 * clockWidget.s)
                                        anchors.verticalCenter: parent.verticalCenter
                                        color: clockWidget.inkDim
                                        opacity: 0.45
                                    }

                                    // Battery / Power
                                    Text {
                                        anchors.verticalCenter: parent.verticalCenter
                                        text: {
                                            if (Battery.available) {
                                                return (Battery.isCharging ? "\u26a1 " : "") + Math.round(Battery.percentage * 100) + "%";
                                            }
                                            return "AC";
                                        }
                                        color: clockWidget.ink
                                        font.family: Appearance.font.family.mono
                                        font.pixelSize: Math.round(12 * clockWidget.s)
                                        font.weight: Font.Medium
                                        style: Text.Raised
                                        styleColor: Qt.rgba(0, 0, 0, 0.45)
                                    }

                                    // Divider
                                    Rectangle {
                                        width: 1
                                        height: Math.round(12 * clockWidget.s)
                                        anchors.verticalCenter: parent.verticalCenter
                                        color: clockWidget.inkDim
                                        opacity: 0.45
                                    }

                                    // Hardware Vitals (CPU + RAM)
                                    Text {
                                        anchors.verticalCenter: parent.verticalCenter
                                        text: {
                                            const cpu = Math.round((ResourceUsage.cpuUsage || 0) * 100) + "% CPU";
                                            const ram = ((ResourceUsage.memoryUsed || 0) / (1024 * 1024)).toFixed(1) + "G";
                                            return cpu + " \u00b7 " + ram;
                                        }
                                        color: clockWidget.inkSoft
                                        font.family: Appearance.font.family.mono
                                        font.pixelSize: Math.round(12 * clockWidget.s)
                                        font.weight: Font.Normal
                                        style: Text.Raised
                                        styleColor: Qt.rgba(0, 0, 0, 0.45)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
