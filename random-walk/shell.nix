let
  pkgs = import <nixpkgs> { };
  libraries = with pkgs.python313Packages; [ matplotlib random2 ];
in pkgs.mkShell {
  # packages = with pkgs.python312Packages;
  #   [ matplotlib random2 ] ++ [ pkgs.python314 ];
  packages = [ pkgs.python313 ] ++ libraries;

  # shellHook = ''
  #   source ./venv/bin/activate.fish
  # '';
}
