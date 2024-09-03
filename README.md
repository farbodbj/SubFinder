# 📡 SubFinder

SubFinder is an open-source tool designed to combat internet censorship by retrieving, testing, and identifying working V2ray configurations from thousands of subscription links. By running this tool, users can maintain a reliable and regularly updated collection of verified V2ray configs tailored to their specific network, making it a powerful resource in the decentralized fight against the Great Firewall (GFW).

## 🛠️ Features

* **Ease of Use**: With a single command, SubFinder will **fetch**, **test**, and **save** working V2ray links, sorted by download speed, into an `output.txt` file. If you fork this repository, it will also **commit and push** these results automatically. A seamless way to ensure you always have the best performing V2ray configurations at hand.

* **Personalized Subscription**: SubFinder can be run as a cron job to continuously update your V2ray configs. Since the testing is done on your network, the resulting configurations are specifically suited to your ISP, ensuring optimal performance.

* **Broadcasting**: The `broadcast.sh` script allows you to easily share the top 10 performing configs to a Telegram channel of your choice, ensuring others can benefit from your optimized configurations.

* **Automation with Cron Job**: If you have an always-on device, you can set up a cron job using the `add_cronjob.sh` script. This will run the test script daily at 18:36, ensuring your V2ray configs are consistently up-to-date.

## 📃 List of Aggregators

SubFinder currently supports the following V2ray aggregators:

* [V2RayAggregator](https://github.com/mahdibland/V2RayAggregator)
* [telegram-configs-collector](https://github.com/soroushmirzaei/telegram-configs-collector)

Feel free to contribute by extending this list with more aggregators!

## 🚀 Getting Started

To get started with SubFinder, follow these steps:

1. **Clone the Repository**:
    ```sh
    git clone git@github.com:farbodbj/SubFinder.git
    cd SubFinder
    ```

2. **Install Golang**:
    Follow the official [Golang installation guide](https://go.dev/doc/install).

3. **Choose Your Path**:

    - **Set Up Cron Job**: 
        Run `add_cronjob.sh` to start the process immediately. Logs will be displayed as the script runs.
    
    - **Run the Main Script**:
        Execute `test_configs_and_push.sh` to test and push the results. This script adds a daily cron job at 18:36 by default.

4. **Enjoy Reliable V2ray Configs** 😄

## 🤝 Contributing

We welcome contributions! Feel free to open issues, submit pull requests, or suggest features.

### Planned Improvements:

- [ ] Dockerize the project for easier deployment.
- [ ] Filter V2ray links from specific countries like China or Hong Kong (due to widespread IP blocking).
- [ ] Expand the list of supported subscriptions.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## 🌟 Acknowledgments

* A huge thanks to the V2Ray community for their incredible work on [v2ray-core](https://github.com/v2ray/v2ray-core).
* This project leverages a fork of the [LiteSpeedTest](https://github.com/xxf098/LiteSpeedTest) library, with some bugs fixed to suit our needs.