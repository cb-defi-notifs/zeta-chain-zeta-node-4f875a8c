import os
import json
import requests
import subprocess
import logging
import sys
from cosmospy import generate_wallet
import time

class Logger:
  def __init__(self):
    self.log = logging.getLogger()
    self.log.setLevel(logging.INFO)
    self.handler = logging.StreamHandler(sys.stdout)
    self.handler.setLevel(logging.DEBUG)
    self.formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    self.handler.setFormatter(self.formatter)
    self.log.addHandler(self.handler)

class GithubBinaryDownload:
    def __init__(self, github_token, github_owner, github_repo):
        self.github_owner = github_owner
        self.github_repo = github_repo
        self.github_token = github_token
        self.upgrades_file_location = "upgrades.json"

    def download_testing_binaries(self):
        load_upgrades = json.loads(open(self.upgrades_file_location, "r").read())
        for asset_url, binary_name in load_upgrades["binary_versions"]:
            headers = {"Accept": "application/vnd.github+json",
                       "Authorization": f"Bearer {self.github_token}"}
            response = requests.get(asset_url, headers=headers).json()
            for asset in response["assets"]:
                if asset["name"].lower() == binary_name.lower():
                    binary_url = asset["browser_download_url"]
                    asset_id = asset["id"]
                    headers = {"Accept": "application/octet-stream",
                               "Authorization": f"Bearer {self.github_token}",
                               "X-GitHub-Api-Version": "2022-11-28"}
                    url_to_download = f"https://api.github.com/repos/{self.github_owner}/{self.github_repo}/releases/assets/{asset_id}"
                    response = requests.get(url_to_download, stream=True, headers=headers)
                    upgrade_version_name = binary_url.split("/")[7].replace("v.", "v")
                    try:
                        os.makedirs("upgrades/", exist_ok=True)
                        os.makedirs("upgrades/" + upgrade_version_name, exist_ok=True)
                        os.makedirs("upgrades/" + upgrade_version_name + "/bin", exist_ok=True)
                    except Exception as e:
                        print(e)
                    with open("upgrades/" + upgrade_version_name + "/bin/zetacored", "wb") as handle:
                        handle.write(response.content)

class Utilities:
    def __init__(self, go_path):
        self.results = {}
        self.go_path = go_path
        self.logger = None
        self.NODE = "http://127.0.0.1:26657"
        self.MONIKER = "zeta"
        self.CHAIN_ID = "test_1001-1"
        self.CONTAINER_ID = None

    def run_command(self, cmd):
        COMMAND_PREFIX = "export PATH="+self.go_path+":${PATH} && "
        cmd = COMMAND_PREFIX + cmd
        result = subprocess.run(cmd, stdout=subprocess.PIPE, shell=True)
        result_output = result.stdout.decode('utf-8')
        return result_output

    def run_command_all_output(self, cmd):
        COMMAND_PREFIX = "export PATH="+self.go_path+":${PATH} && "
        cmd = COMMAND_PREFIX + cmd
        result = subprocess.run(cmd, stdout=subprocess.PIPE, shell=True)
        result_output = result.stdout.decode('utf-8')
        try:
            error_output = result.stderr.decode('utf-8')
        except:
            error_output = ""
        return result_output, error_output

    def generate_wallet(self):
        wallet = generate_wallet()
        #self.address = wallet["address"]
        self.mnemonic = wallet["seed"]
        self.derivation_path = wallet["derivation_path"]
        return self

    def get_proposal_id(self):
        try:
            QUERY_GOV_PROPOSAL = f"""zetacored query gov proposals --output json --node {self.NODE}"""
            GOV_PROPOSALS = json.loads(self.run_command(QUERY_GOV_PROPOSAL))
            for proposal in GOV_PROPOSALS["proposals"]:
                PROPOSAL_ID = proposal["proposal_id"]
            return PROPOSAL_ID
        except Exception as e:
            self.logger.error(e)
            return 1

    def raise_governance_proposal(self,VERSION,BLOCK_TIME_SECONDS, PROPOSAL_TIME_SECONDS, UPGRADE_INFO):
        try:
            self.CURRENT_HEIGHT = requests.get(f"{self.NODE}/status").json()["result"]["sync_info"]["latest_block_height"]
            self.UPGRADE_HEIGHT = str(int(self.CURRENT_HEIGHT) + (PROPOSAL_TIME_SECONDS / BLOCK_TIME_SECONDS)).split(".")[0]
            GOV_PROPOSAL = f"""zetacored tx gov submit-proposal software-upgrade "{VERSION}" \
                --from "{self.MONIKER}" \
                --deposit 10000000000000000000azeta \
                --upgrade-height "{self.UPGRADE_HEIGHT}" \
                --upgrade-info "{UPGRADE_INFO}" \
                --title "{VERSION}" \
                --description "Zeta Release {VERSION}" \
                --chain-id "{self.CHAIN_ID}" \
                --node "{self.NODE}" \
                --keyring-backend test \
                --fees 20000azeta \
                -y"""
            self.logger.info(GOV_PROPOSAL)
            results_output = self.run_command(GOV_PROPOSAL)
            self.logger.info(results_output)
            TX_HASH = results_output.split("\n")[12].split(":")[1].strip()
            self.logger.info(TX_HASH)
        except Exception as e:
            self.logger.error(str(e))
            self.logger.error("There was issue rasing proposal the old way swtich to new way and see if that helps.")
            self.CURRENT_HEIGHT = requests.get(f"{self.NODE}/status").json()["result"]["sync_info"][
                "latest_block_height"]
            self.UPGRADE_HEIGHT = \
            str(int(self.CURRENT_HEIGHT) + (PROPOSAL_TIME_SECONDS / BLOCK_TIME_SECONDS)).split(".")[0]
            GOV_PROPOSAL = f"""zetacored tx gov submit-legacy-proposal software-upgrade "{VERSION}" \
                --from "{self.MONIKER}" \
                --deposit 10000000000000000000azeta \
                --upgrade-height "{self.UPGRADE_HEIGHT}" \
                --upgrade-info "{UPGRADE_INFO}" \
                --title "{VERSION}" \
                --description "Zeta Release {VERSION}" \
                --chain-id "{self.CHAIN_ID}" \
                --node "{self.NODE}" \
                --keyring-backend test \
                --fees 20000azeta \
                -y \
                --no-validate"""
            self.logger.info(GOV_PROPOSAL)
            results_output = self.run_command(GOV_PROPOSAL)
            self.logger.info(results_output)
            TX_HASH = results_output.split("\n")[12].split(":")[1].strip()
            self.logger.info(TX_HASH)

        return TX_HASH, self

    def raise_governance_vote(self, PROPOSAL_ID):
        VOTE_PROPOSAL=f"""zetacored tx gov vote "{PROPOSAL_ID}" yes \
            --from {self.MONIKER} \
            --keyring-backend test \
            --chain-id {self.CHAIN_ID} \
            --node {self.NODE} \
            --fees 20000azeta \
            -y"""
        self.logger.info(VOTE_PROPOSAL)
        results_output = self.run_command(VOTE_PROPOSAL)
        return results_output

    def load_key(self):
        LOAD_KEY = f"echo {self.mnemonic} | zetacored keys add {self.MONIKER} --keyring-backend test --recover"
        DELETE_KEY = f"yes | zetacored keys delete {self.MONIKER} --keyring-backend test"
        self.run_command(DELETE_KEY)
        self.address = self.run_command(LOAD_KEY).split("\n")[2].split(":")[1].strip()

    def kill_docker_containers(self):
        docker_containers_command_output = self.run_command("docker ps | grep -v COMMAND").split("\n")
        self.logger.info(docker_containers_command_output)
        for contianer in docker_containers_command_output:
            try:
                container_id = contianer.split()[0]
                kill_docker_container = self.run_command(f"docker kill {container_id}")
                self.logger.info (kill_docker_container)
            except Exception as e:
                self.logger.error(str(e))

    def version_check(self, END_VERSION):
        version_check_request = requests.get(self.NODE+"/abci_info").json()
        VERSION_CHECK = version_check_request["result"]["response"]["version"]
        self.logger.info(f"END_VERSION: {END_VERSION}")
        self.logger.info(f"CURRENT_VERSION: {VERSION_CHECK}")
        if VERSION_CHECK == END_VERSION:
            return True
        else:
            return False

    def current_block(self):
        block_height_request = requests.get(self.NODE+"/status").json()
        LATEST_BLOCK = block_height_request["result"]["sync_info"]["latest_block_height"]
        return int(LATEST_BLOCK)

    def docker_ps(self):
        self.run_command(f'docker ps')

    def build_docker_image(self, docker_file_location):
        self.logger.info("Build Docker Image")
        docker_build_output = self.run_command(f'docker build -t local/upgrade-test:latest {docker_file_location}')
        self.logger.info(docker_build_output)
        docker_image_list = self.run_command("docker image list")
        self.logger.info(docker_image_list)

    def start_docker_container(self, GAS_PRICES,
                               DAEMON_HOME,
                               DAEMON_NAME,
                               DENOM,
                               DAEMON_ALLOW_DOWNLOAD_BINARIES,
                               DAEMON_RESTART_AFTER_UPGRADE,
                               EXTERNAL_IP,
                               STARTING_VERSION,
                               VOTING_PERIOD,
                               LOG_LEVEL,
                               UNSAFE_SKIP_BACKUP,
                               CLIENT_DAEMON_NAME,
                               CLIENT_DAEMON_ARGS,
                               CLIENT_SKIP_UPGRADE,
                               CLIENT_START_PROCESS):
        DOCKER_ENVS = f""" 
 -e MONIKER='{self.MONIKER}'
 -e GAS_PRICES='{GAS_PRICES}'
 -e DAEMON_HOME='{DAEMON_HOME}'
 -e DAEMON_NAME='{DAEMON_NAME}'
 -e DENOM='{DENOM}'
 -e DAEMON_ALLOW_DOWNLOAD_BINARIES='{DAEMON_ALLOW_DOWNLOAD_BINARIES}'
 -e DAEMON_RESTART_AFTER_UPGRADE='{DAEMON_RESTART_AFTER_UPGRADE}'
 -e CHAIN_ID='{self.CHAIN_ID}'
 -e EXTERNAL_IP='{EXTERNAL_IP}'
 -e STARTING_VERSION='{STARTING_VERSION}'
 -e VOTING_PERIOD='{VOTING_PERIOD}s'
 -e LOG_LEVEL='{LOG_LEVEL}'
 -e UNSAFE_SKIP_BACKUP='{UNSAFE_SKIP_BACKUP}'
 -e ZETA_MNEMONIC='{self.mnemonic}'
 -e CLIENT_DAEMON_NAME='{CLIENT_DAEMON_NAME}'
 -e CLIENT_DAEMON_ARGS='{CLIENT_DAEMON_ARGS}'
 -e CLIENT_SKIP_UPGRADE='{CLIENT_SKIP_UPGRADE}'
 -e CLIENT_START_PROCESS='{CLIENT_START_PROCESS}' 
        """
        DOCKER_ENVS = DOCKER_ENVS.replace("\n", " ")
        self.logger.info("kill running containers.")
        self.kill_docker_containers()
        self.logger.info("Start local network contianer.")
        docker_command = f'docker run {DOCKER_ENVS} -d -p 26657:26657 local/upgrade-test:latest'
        self.logger.info(docker_command)
        self.run_command(docker_command)
        container_id = self.run_command("docker ps | grep -v COMMAND | cut -d ' ' -f 1 | tr -d ' '")
        self.CONTAINER_ID = container_id
        time.sleep(3)
        self.logger.info("ContainerID")
        self.logger.info(container_id)
        time.sleep(60)
        docker_logs,error_output = self.run_command_all_output(f'docker logs {container_id}')
        self.logger.info(docker_logs)
        self.logger.error(error_output)
        docker_ps = self.run_command("docker ps")
        self.logger.info(docker_ps)
        time.sleep(60)
        return True