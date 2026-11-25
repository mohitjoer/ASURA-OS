import QtQuick
import org.kde.kirigami as Kirigami
import qs.services
import qs.modules.common

Kirigami.Icon {
    id: root
    required property string iconName
    property bool separateLightDark: false
    property bool tryCustomIcon: true
    
    property real implicitSize: 26
    implicitWidth: implicitSize
    implicitHeight: implicitSize

    animated: true
    roundToIconSize: false
    fallback: "application-x-executable"
    source: {
        if (!root.iconName || root.iconName.length === 0) return "application-x-executable";
        if (root.tryCustomIcon) {
            return `${Looks.iconsPath}/${root.iconName}${!root.separateLightDark ? "" : Looks.dark ? "-dark" : "-light"}.svg`;
        }
        let guess = AppSearch.guessIcon(root.iconName);
        if (guess && guess !== "image-missing" && AppSearch.iconExists(guess)) return guess;
        if (AppSearch.iconExists(root.iconName)) return root.iconName;
        return "application-x-executable";
    }

    color: Looks.colors.fg
}
