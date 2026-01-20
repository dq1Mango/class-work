let pkgs = import <nixpkgs> { };
in pkgs.mkShell {
  packages = with pkgs.python312Packages; [ matplotlib random2 ];

  # shellHook = ''
  #   source ./venv/bin/activate.fish
  # '';
}
