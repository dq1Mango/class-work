import pickle

import matplotlib as plt


def load_data(path):
    with open(path, "rb") as file:
        return pickle.load(file)


def show_graphs(sucesses):

    points = len(sucesses)

    first_derivative = [
        (sucesses[i + 1] - sucesses[i]) / step for i in range(points - 1)
    ]
    second_derivative = [
        (first_derivative[i + 1] - first_derivative[i]) / step
        for i in range(points - 2)
    ]

    infelctions = []

    for i in range(len(second_derivative) - 1):
        first = sign(second_derivative[i])
        second = sign(second_derivative[i + 1])

        if first == 0:
            infelctions.append(float(probs[i]))

        if first == 1 and second == -1:
            infelctions.append(float(probs[i]))

    print(infelctions)

    fig, ax = plt.subplots()
    ax.set_title("Density vs Percolation")
    ax.set_xlabel("Filled %")
    ax.set_ylabel("Percolation %")
    ax.plot(probs, sucesses)
    plt.show()

    fig.savefig("figure")
