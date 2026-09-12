# run.py

import argparse
from app import create_app

app = create_app(
    secret="foo"
)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Flask app")
    parser.add_argument(
        "--port",
        type=int,
        default=6742,
        help="Select port"
    )

    args = parser.parse_args()

    app.run(
        host="127.0.0.1",
        port=args.port,
        debug=True,
    )