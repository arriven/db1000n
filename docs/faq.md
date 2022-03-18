# FAQ

???+ faq "Where can I find advanced documentation?"

    [Here](/db1000n/advanced-docs/advanced-and-devs/)

---

???+ faq "I installed `db1000n` but it's not working properly. What to do?"

    Create [Issue](https://github.com/Arriven/db1000n/issues) and community will help you with solving a problem

---

???+ faq "I'm **not** a developer, how can I help to project?"

    - Share information about `db1000n` in social media, with your friends and colleagues
    - Run `db1000n` on every possible platform (local machine, public clouds, Docker, Kubernetes, etc)
    - Create [Issues](https://github.com/Arriven/db1000n/issues) or
      [Pull Requests](https://github.com/Arriven/db1000n/pulls)
      if you found any bugs, missed documentation, misspells, etc

---

???+ faq "I'm a developer, how can I help to project?"

    - Check [Issues](https://github.com/Arriven/db1000n/issues) to help with important tasks
    - Check our codebase and make [PRs](https://github.com/Arriven/db1000n/pulls)
    - Test an app on different platforms and report bugs or feature requests

---

???+ faq "When I run `db1000n` I see that it generates low amount of traffic. Isn't that bad?"

    ???+ info "it's okay"

        The app is configurable to generate set amount of traffic (controlled by the number
        of targets, their type, and attack interval for each of them).
        The main reason it works that way is because there are two main types of ddos:

        - Straightforward load generation (easy to implement, easy to defend from) - as effective
          as the amount of raw traffic you can generate

        - Actual denial of service that aims to remain as undetected as possible by simulating plausible
          traffic and hitting concrete vulnerabilities in the target (or targets). This type of ddos doesn't
          require a lot of traffic and thus is mostly limited by the amount of clients generating this type
          of load (or rather unique IPs)

---

???+ faq "Should I care about costs if I run an app in public cloud?"

    ???+ info "[Yes](https://github.com/Arriven/db1000n/issues/153)"

        Cloud providers could charge a huge amount of money not only for compute resources but for traffic as well.
        If you run an app in the cloud please control your billing
        (if you use Docker, ensure that use advanced image: `ghcr.io/arriven/db1000n-advanced`)

---

???+ faq "Can I leave the site for the night?"

    Yes, you can. I personally leave the browser on overnight and it works fine.

---

???+ faq "How can I make sure that the computer does not go to sleep while the site is running?"

    To do this, you need to install a program which keeps the screen turned off. Instructions for different operating systems below:

    - I have Windows: Caffeinated ([download](https://www.microsoft.com/en-us/p/windows-caffeinated/9pbvhhsn78bl?activetab=pivot:overviewtab))
    - I have Mac OS: Amphetamine ([download](https://apps.apple.com/us/app/amphetamine/id937984704?mt=12))

---

???+ faq "How can I make it more effective?"

    When selecting a VPN country, choose "friendly" countries for Russia:

    - Russia
    - Belarus
    - Kazakhstan

    ???+ warning "Objective"

        Generate traffic from within, from their internal address.
        Russia can close its propaganda to the world,
        but they cannot close it to their citizens

---
