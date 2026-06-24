#include <QApplication>
#include "mainwindow.h"

int main(int argc, char *argv[]) {
    // ponytail: FFmpeg backend handles pause correctly; WMF keeps audio playing on pause
    qputenv("QT_MEDIA_BACKEND", "ffmpeg");
    QApplication app(argc, argv);
    app.setApplicationName("ber-client");
    app.setOrganizationName("ber");
    app.setApplicationVersion("1.0.0");

    MainWindow window;
    window.show();

    return app.exec();
}
