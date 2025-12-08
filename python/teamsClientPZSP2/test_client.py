import client

def main():
    teams_client = client.TeamsClient()
    team_ref = "pzsp2"

    channels = teams_client.list_channels(team_ref)
    print("Channels:", channels)




if __name__ == "__main__":
    main()