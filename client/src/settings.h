#pragma once

#include <QObject>
#include <QSettings>
#include <QString>

class Settings : public QObject {
    Q_OBJECT

public:
    explicit Settings(QObject *parent = nullptr);

    QString lastServer() const;
    void setLastServer(const QString &server);

    int volume() const;
    void setVolume(int volume);

private:
    QSettings m_settings;
};
