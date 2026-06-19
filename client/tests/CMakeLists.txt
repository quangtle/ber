include(FetchContent)
FetchContent_Declare(
    Catch2
    GIT_REPOSITORY https://github.com/catchorg/Catch2.git
    GIT_TAG v3.5.4
)
FetchContent_MakeAvailable(Catch2)

add_executable(ber-client-tests
    main.cpp
    test_apiclient.cpp
    test_settings.cpp
    ${CMAKE_SOURCE_DIR}/src/apiclient.cpp
    ${CMAKE_SOURCE_DIR}/src/settings.cpp
)

target_link_libraries(ber-client-tests PRIVATE
    Catch2::Catch2WithMain
    Qt6::Core
    Qt6::Network
)

target_include_directories(ber-client-tests PRIVATE ${CMAKE_SOURCE_DIR}/src)

include(CTest)

add_test(NAME ber-client-tests COMMAND ber-client-tests)
if(WIN32)
    set_tests_properties(ber-client-tests PROPERTIES
        ENVIRONMENT "PATH=$ENV{PATH}"
    )
endif()
