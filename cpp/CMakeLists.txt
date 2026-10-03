cmake_minimum_required(VERSION 3.16)
project(texas_algorithm CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Threads REQUIRED)

add_library(texas_algorithm
    poke.cpp
    texas_card_util.cpp
    gen_util.cpp
    sorter.cpp
    gen_opt_util.cpp
    gen_extra_util.cpp
    gen_extra_opt_util.cpp
    gen_trans_util.cpp
    gen_trans_opt_util.cpp
    gen_hand_util.cpp
    texas_algorithm_util.cpp
)
target_include_directories(texas_algorithm PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(texas_algorithm PUBLIC Threads::Threads)

add_executable(texas_algorithm_demo cmd/texas_algorithm/main.cpp)
target_link_libraries(texas_algorithm_demo PRIVATE texas_algorithm)

enable_testing()

add_executable(texas_card_util_test tests/texas_card_util_test.cpp)
target_link_libraries(texas_card_util_test PRIVATE texas_algorithm)
add_test(NAME texas_card_util_test COMMAND texas_card_util_test
         WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR})

add_executable(texas_algorithm_util_test tests/texas_algorithm_util_test.cpp)
target_link_libraries(texas_algorithm_util_test PRIVATE texas_algorithm)
add_test(NAME texas_algorithm_util_test COMMAND texas_algorithm_util_test
         WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR})
