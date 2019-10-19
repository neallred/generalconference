from general_conference.main import denest

def test_denest():
    assert denest([[1],[2],[3]]) == [1,2,3]

