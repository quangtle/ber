#include "playerwidget.h"
#include "videoplayer.h"
#include "apiclient.h"

#include <QHBoxLayout>

PlayerWidget::PlayerWidget(ApiClient *client, QWidget *parent)
    : QWidget(parent)
    , m_apiClient(client)
    , m_videoPlayer(new VideoPlayer(this))
    , m_backBtn(new QPushButton("← Library"))
    , m_playPauseBtn(new QPushButton("Pause"))
    , m_titleLabel(new QLabel())
{
    setupUi();
}

void PlayerWidget::setupUi() {
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);

    auto *toolbar = new QWidget();
    auto *tbLayout = new QHBoxLayout(toolbar);
    tbLayout->setContentsMargins(8, 4, 8, 4);
    tbLayout->addWidget(m_backBtn);
    tbLayout->addWidget(m_titleLabel, 1);
    tbLayout->addWidget(m_playPauseBtn);

    layout->addWidget(toolbar);
    layout->addWidget(m_videoPlayer, 1);

    connect(m_backBtn, &QPushButton::clicked, this, &PlayerWidget::backToLibrary);
    connect(m_playPauseBtn, &QPushButton::clicked, m_videoPlayer, &VideoPlayer::togglePlayPause);
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
