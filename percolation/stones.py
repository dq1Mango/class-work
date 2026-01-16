import random

import hoshen_kopelman

# p = 0.2


def reset_color():
    print(f"\x1b[37m")


def known_gen_grid():
    return [[0, 0, 0], [1, 1, 1], [0, 0, 0]]


def gen_grid(size, p):
    grid = [[0 for i in range(size)] for i in range(size)]

    for i in range(size):
        for j in range(size):
            if random.random() <= p:
                grid[i][j] = 1

    return grid


def print_grid(grid):
    block = "██"
    for row in grid:
        # print("[", end = '')
        for stone in row:
            if stone == 0:
                print("\x1b[30m", end="")
            else:
                print(f"\x1b[38;5;{stone-2}m", end="")

            print(f"{block}", end="")
            # print(f"{stone} ", end="")

        print()

    print("\x1b[0m", end="")


def find_path(grid, path):
    # print(path)
    # print()
    final = len(grid) - 1
    (past_x, past_y) = path[-1][0], path[-1][1]
    if (final == past_x and path[0][0] == 0) or (final == past_y and path[0][1] == 0):
        # print("we won")
        return [path]

    found_paths = []
    for delta_x, delta_y in [(0, 1), (1, 0), (-1, 0), (0, -1)]:
        new_x, new_y = past_x + delta_x, past_y + delta_y

        if (new_x < 0 or new_x >= len(grid)) or (new_y < 0 or new_y >= len(grid)):
            continue
        # print("heres the new coords:")
        # print(new_x, new_y)

        if len(path) > 1:
            if (new_x, new_y) == path[-2]:
                # print("why dont u skip that")
                continue

        if (new_x, new_y) in path:
            # print("this works right?")
            # print(new_x, new_y)
            return 2

        if grid[new_x][new_y] == 1:
            result = find_path(grid, path + [(new_x, new_y)])

            # if result == 2:
            #     return 2

            if result != 1 and result != 2:
                # print(f"adding result: {result}")
                found_paths += result

    # print("ok we done")
    # print(found_paths)
    # if len(found_path) > 0:
    #     # print("look we found one")
    #     return found_path
    # else:
    #    return 1

    return found_paths

def check_percolated(grid):
    size = len(grid)

    clusters_left = []
    clusters_up = []
    clusters_down = []
    clusters_right = []

    for i in range(size):

        if grid[i][0] != 0:
            if grid[i][0] not in clusters_left:
                clusters_left.append(grid[i][0])

        if grid[i][size - 1] != 0:
            if grid[i][size - 1] not in clusters_right:
                clusters_right.append(grid[i][size - 1])

        if grid[0][i] != 0:
            if grid[0][i] not in clusters_up:
                clusters_up.append(grid[0][i])

        if grid[size - 1][i] != 0:
            if grid[size - 1][i] not in clusters_down:
                clusters_down.append(grid[size - 1][i])

    # print(f"left: {clusters_left}")
    # print(f"right: {clusters_right}")
    # print(f"up: {clusters_up}")
    # print(f"down: {clusters_down}")

    for cluster in clusters_left:
        if cluster in clusters_right:
            return True

    for clusters in clusters_up:
        if cluster in clusters_down:
            return True


    return False



size = 20
p = 0.5

no_paths = 0
percolated = 0

trials = 100

for i in range(trials):
    grid = gen_grid(size, p)

    grid = hoshen_kopelman.countClusters(grid)

    # print_grid(grid)
    
    check = check_percolated(grid)

    if check:
        # print("we percolated!!!")
        percolated += 1
    else:
        # print("could not get across")
        no_paths += 1

    # print()


# old bad way
# for p in range(50, 51):
# for i in range(trials):
#     print(i)
#     # p = p / 100
#     # print(f"\x1b[37mp value of {p}:")
#     grid = gen_grid(size, p)
#     print_grid(grid)
#     # grid = known_gen_grid()
#
#     paths = []
#     failed = False
#
#     # find_path(grid, [(0, 0)])
#     for x in range(1, size):
#         if grid[x][0] == 0:
#             continue
#
#         result = find_path(grid, [(x, 0)])
#         if len(result) > 0:
#             paths += result
#
#         # if result == 2:
#         #     failed = True
#         # if result == 1:
#         #     continue
#         # else:
#         #     paths += [result]
#
#     for y in range(1, size):
#         if grid[0][y] == 0:
#             continue
#
#         result = find_path(grid, [(0, y)])
#         if len(result) > 0:
#             paths += result
#         # if result == 2:
#         #     failed = True
#         # if result == 1:
#         #     continue
#         # else:
#         #     paths += [result]
#
#     if len(paths) == 1:
#         reset_color()
#         print("omg no way one worked!!!")
#         print_grid(grid)
#         reset_color()
#         print(paths[0])
#         print()
#         acceptable += 1
#
#     elif len(paths) > 1:
#         print("too many")
#         multiple_paths += 1
#
#     else:
#         print("none at all")
#         no_paths += 1


print(f"\x1b[37mp value of {p}:")
print(f"total trials:\t{trials}")
print(f"no paths: \t{no_paths} - {no_paths/trials*100}%")
print(f"percolated 👍: \t{percolated} - {percolated/trials*100}%")
