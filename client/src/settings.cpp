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
