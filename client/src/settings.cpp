#include "settings.h"

Settings::Settings(QObject *parent)
    : QObject(parent)
    , m_settings("ber", "ber-client")
{
}

QString Settings::lastServer() const {
    return m_settings.value("server/last", "").toString();
}

void Settings::setLastServer(const QString &server) {
    m_settings.setValue("server/last", server);
}

int Settings::volume() const {
    return m_settings.value("player/volume", 100).toInt();
}

void Settings::setVolume(int volume) {
    m_settings.setValue("player/volume", volume);
}
