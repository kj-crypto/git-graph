import tkinter as tk
import json
from tkinter import Canvas, Frame, Scrollbar, HORIZONTAL, VERTICAL, BOTH, X, Y, BOTTOM, RIGHT, LEFT


def calculate_bezier_point(t, P0, P1, P2, P3):
    """Compute a point along a cubic Bézier curve at parameter t."""
    x = (1-t)**3 * P0[0] + 3*(1-t)**2 * t * P1[0] + 3*(1-t) * t**2 * P2[0] + t**3 * P3[0]
    y = (1-t)**3 * P0[1] + 3*(1-t)**2 * t * P1[1] + 3*(1-t) * t**2 * P2[1] + t**3 * P3[1]
    return (x, y)


def calc_quad_bezier(t, P0, P1, P2):
    """Compute a point along a quadratic Bézier curve at parameter t."""
    x = (1 - t)**2 * P0[0] + 2*(1 - t)*t * P1[0] + t**2 * P2[0]
    y = (1 - t)**2 * P0[1] + 2*(1 - t)*t * P1[1] + t**2 * P2[1]
    return (x, y)


def draw_connection(canvas, c1, c2, R, steps=50):
    colors = ['blue', 'green', 'purple', 'orange', 'pink', 'cyan']
    x1, y1 = c1
    x2, y2 = c2

    if x1 == x2:
        color = colors[int(x1/R/2.5) % len(colors)]
        canvas.create_line(x1, y1 + R, x2, y2 - R, fill=color, width=4)
        return

    if x2 > x1:
        P0 = (x1 + R, y1)
        P1 = (x2, y1)
        P2 = (x2, y2 -R)
    else:
        P0 = (x1, y1+ R)
        P1 = (x1, y2)
        P2 = (x2 + R, y2)

    points = []
    for i in range(steps + 1):
        t = i / steps
        x, y = calc_quad_bezier(t, P0, P1, P2)
        points.extend([x, y])

    canvas.create_line(points, fill='black', width=2, smooth=True)


class Commit:
    def __init__(self, _hash, x_pos, y_pos, parents, message):
        self.hash = _hash
        self.x_pos = x_pos
        self.y_pos = y_pos
        self.parents = parents
        self.message = message

    @classmethod
    def from_file(cls, file_path):
        with open(file_path, 'r') as f:
            commits = json.load(f)

        return {commit['hash']: Commit(commit['hash'], commit['x_pos'], commit['y_pos'], commit['parents'], commit['message']) for commit in commits}


class GitGraphVisualizer:
    def __init__(self, tk_root, commits: dict[str, Commit], R=30):
        self.canvas = None
        self.commits = commits
        
        self.R = R
        self.X_shift = 1.1 * R
        self.Y_shift = 1.1 * R
        self.vertical_spacing = 2.5 * R
        self.horizontal_spacing = 2.5 * R
        self._init_window(tk_root)

    def _init_window(self, tk_root):
        frame = Frame(tk_root)
        frame.pack(fill=BOTH, expand=True)

        h_scrollbar = Scrollbar(frame, orient=HORIZONTAL, width=10)
        v_scrollbar = Scrollbar(frame, orient=VERTICAL, width=10)

        self.canvas = Canvas(frame, width=100, height=100, 
            xscrollcommand=h_scrollbar.set,
            yscrollcommand=v_scrollbar.set, 
            highlightthickness=0, borderwidth=0,
        )
        h_scrollbar.config(command=self.canvas.xview)
        v_scrollbar.config(command=self.canvas.yview)

        h_scrollbar.pack(side=BOTTOM, fill=X)
        v_scrollbar.pack(side=RIGHT, fill=Y)

        self.canvas.pack(side=LEFT, fill=BOTH, expand=True)
        max_x = max(self.commits.values(), key=lambda commit: commit.x_pos).x_pos
        max_y = max(self.commits.values(), key=lambda commit: commit.y_pos).y_pos
        scroll_bar_paddings = 15
        width = max_x * self.horizontal_spacing + 2 * self.X_shift
        width = min(width + scroll_bar_paddings, tk_root.winfo_screenwidth() / 2)
        height = max_y * self.vertical_spacing + 2 * self.Y_shift
        height = min(height + scroll_bar_paddings, tk_root.winfo_screenheight())

        x = (tk_root.winfo_screenwidth() - width) // 2
        y = (tk_root.winfo_screenheight() - height) // 2

        self.canvas.config(width=width - scroll_bar_paddings, height=height - scroll_bar_paddings)
        tk_root.geometry(f"{int(width)}x{int(height)}+{int(x)}+{int(y)}")

    def get_position(self, commit: Commit):
        return (commit.x_pos * self.horizontal_spacing + self.X_shift,
                commit.y_pos * self.vertical_spacing + self.Y_shift)

    def draw_ellipse(self, commit: Commit):
        x, y = self.get_position(commit)
        self.canvas.create_oval(
            x - self.R, y - self.R, x + self.R, y + self.R,
            fill='white', outline='white'
        )
        if commit.hash.startswith("dummy_"):
            text = "d_" + commit.hash.split("_")[1]
            self.canvas.create_text(
                x, y,
                text=text,
                font=('Arial', 10, 'bold'),
        )
        else:
            self.canvas.create_text(
                x, y - 5,
                text=commit.hash[:6],
                font=('Arial', 10)
            )
            message = commit.message if len(commit.message) < 12 else commit.message[:12]
            self.canvas.create_text(
                x, y + 7,
                text=message,
                font=('Arial', 10, 'bold'),
            )

    def draw_graph(self):
        for commit in self.commits.values():
            self.draw_ellipse(commit)

        for commit in self.commits.values():
            for parent_hash in commit.parents:
                parent = self.commits.get(parent_hash)
                if parent is None:
                    continue
                draw_connection(self.canvas, self.get_position(commit), self.get_position(parent), self.R)


def visualize_graph(commits: list[Commit]):
    root = tk.Tk()
    root.title("Git Graph Visualizer")
    visualizer = GitGraphVisualizer(root, commits)
    visualizer.draw_graph()
    root.mainloop()


def visualize_from_file(file_path: str):
    commits = Commit.from_file(file_path)
    visualize_graph(commits)


if __name__ == "__main__":
    visualize_from_file("commit_positions.json")

