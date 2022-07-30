custom_manager "paru" {
  cmd = "paru"
  flags = ["--noconfirm"]
  action "clean" {
    flags = ["-Sc", "--cleanafter"]
  }
  action "install" {
    flags = ["-S", "--needed"]
  }
  action "remove" {
    flags = ["-Rcs"]
  }
  action "refresh" {
    flags = ["-Syy"]
  }
  action "update" {
    flags = ["-Su"]
  }
}

custom_manager "pacman" {
  cmd = "pacman"
  flags = ["--noconfirm"]
  action "clean" {
    flags = ["-Sc"]
  }
  action "install" {
    flags = ["-S", "--needed"]
  }
  action "remove" {
    flags = ["-Rcs"]
  }
  action "refresh" {
    flags = ["-Syy"]
  }
  action "update" {
    flags = ["-Su"]
  }
}

custom_manager "apk" {
  cmd = "apk"
  flags = ["--no-cache"]
  action "clean" {
    flags = ["-Sc"]
  }
  action "install" {
    flags = ["add"]
  }
  action "remove" {
    flags = ["del", "--rdepends"]
  }
  action "refresh" {
    flags = ["update"]
  }
  action "update" {
    flags = ["upgrade"]
  }
}


custom_manager "apt" {
  cmd = "apt-get"
  flags = "-y"
  action "clean" {
    flags = ["clean"]
  }
  action "install" {
    flags = ["install"]
  }
  action "remove" {
    flags = ["remove", "--auto-remove"]
  }
  action "refresh" {
    flags = ["update"]
  }
  action "update" {
    flags = ["dist-upgrade"]
  }
}