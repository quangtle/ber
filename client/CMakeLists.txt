cmake_minimum_required(VERSION 3.28)
project(ber-client VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

find_package(Qt6 REQUIRED COMPONENTS Core Gui Widgets Multimedia MultimediaWidgets Network)

set(SOURCES
    src/main.cpp
    src/mainwindow.cpp
    src/connectiondialog.cpp
    src/libraryview.cpp
    src/playerwidget.cpp
    src/videoplayer.cpp
    src/apiclient.cpp
    src/settings.cpp
)

set(HEADERS
    src/mainwindow.h
    src/connectiondialog.h
    src/libraryview.h
    src/playerwidget.h
    src/videoplayer.h
    src/apiclient.h
    src/settings.h
)

set(RESOURCES
    resources/ber.qrc
)

add_executable(ber-client ${SOURCES} ${HEADERS} ${RESOURCES})

target_link_libraries(ber-client PRIVATE
    Qt6::Core
    Qt6::Gui
    Qt6::Widgets
    Qt6::Multimedia
    Qt6::MultimediaWidgets
    Qt6::Network
)

target_include_directories(ber-client PRIVATE src)

if(WIN32)
    set_target_properties(ber-client PROPERTIES WIN32_EXECUTABLE TRUE)

    find_program(WINDEPLOYQT windeployqt
        PATHS "${Qt6_DIR}/../../../bin"
        NO_DEFAULT_PATH
    )
    if(WINDEPLOYQT)
        add_custom_command(TARGET ber-client POST_BUILD
            COMMAND ${WINDEPLOYQT} $<TARGET_FILE:ber-client>
        )
    endif()
endif()

install(TARGETS ber-client RUNTIME DESTINATION .)

# Tests
option(BUILD_TESTS "Build client tests" ON)
if(BUILD_TESTS)
    enable_testing()
    add_subdirectory(tests)
endif()
