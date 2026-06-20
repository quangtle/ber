#include "playerwidget.h"
#include "videoplayer.h"
#include "apiclient.h"

#include <QVBoxLayout>
#include <QHBoxLayout>

PlayerWidget::PlayerWidget(ApiClient *client, QWidget *parent)
    : QWidget(parent)
    , m_apiClient(client)
    , m_videoPlayer(new VideoPlayer(this))
    , m_backBtn(new QPushButton("← Library"))
    , m_titleLabel(new QLabel())
    , m_toolbar(new QWidget(this))
    , m_hideTimer(new QTimer(this))
{
    setupUi();

    m_hideTimer->setSingleShot(true);
    m_hideTimer->setInterval(5000);
    connect(m_hideTimer, &QTimer::timeout, this, [this]() { m_toolbar->hide(); });

    connect(m_videoPlayer, &VideoPlayer::mouseActivity, this, &PlayerWidget::resetHideTimer);
    connect(m_videoPlayer, &VideoPlayer::fullscreenToggled, this, &PlayerWidget::fullscreenToggled);
}

void PlayerWidget::setupUi() {
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);

    auto *tbLayout = new QHBoxLayout(m_toolbar);
    tbLayout->setContentsMargins(8, 4, 8, 4);
    tbLayout->addWidget(m_backBtn);
    tbLayout->addWidget(m_titleLabel, 1);

    layout->addWidget(m_toolbar);
    layout->addWidget(m_videoPlayer, 1);

    connect(m_backBtn, &QPushButton::clicked, this, &PlayerWidget::backToLibrary);
}

void PlayerWidget::playVideo(const QString &videoId) {
    m_currentVideoId = videoId;
    QString streamUrl = m_apiClient->streamUrl(videoId);
    m_videoPlayer->load(streamUrl);
    m_titleLabel->setText("Loading...");

    m_apiClient->fetchVideoInfo(videoId, [this](const QJsonObject &info) {
        m_titleLabel->setText(info["title"].toString());
    });
}

void PlayerWidget::resetHideTimer() {
    m_toolbar->show();
    m_hideTimer->start();
}
