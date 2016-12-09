*** Settings ***
| Library | Process
| Library | OperatingSystem
#| Library | SSHLibrary

*** Test cases ***
| Test1
| | ${result1} | ${output1}= | Run and Return RC and Output | curl -vv -k -X POST -d "{\"username\": \"vishnu\", \"email\": \"vishnu@cisco.com\", \"password\": \"vishnu123\"}" https://localhost:8443/user
| | BuiltIn.Should Be Equal As Integers |  ${result1} | 0
| | Should Contain | ${output1} | User
| Test2
| | ${result2} | ${output2}= | Run and Return RC and Output | curl -vv -k -X PUT -d "{\"billingContact\": \”Vishnu Konepalli\”, \"email\": \"vishnu@cisco.com\", \"address\": \"3550 Cisco Way San Jose CA 95134\", \"zipCode\": \"95134\", \"carLicensePlat\": \"5jdg099\"}" https://localhost:8443/user/vishnu/profile 
| | BuiltIn.Should Be Equal As Integers |  ${result2} | 0
| | Should Contain | ${output2} | 200 
| Test3
| | ${result3} | ${output3}= | Run and Return RC and Output | curl -vv -k -X POST -d "{\"Parkingid\": \"A123\"}" https://localhost:8443/user/vishnu/park 
| | BuiltIn.Should Be Equal As Integers |  ${result3} | 0
| | Should Contain | ${output3} | 200
| Test4
| | ${result4} | ${output4}= | Run and Return RC and Output | curl -vv -k -X PUT -d "{\"shareLicencePlate\": \"Yes\", \"shareParkingDuration\": \"4\",\"shareServiceUsages\": \"Yes\"}" https://localhost:8443/user/vishnu/privacy
| | BuiltIn.Should Be Equal As Integers |  ${result4} | 0
| | Should Contain | ${output4} | 200
| Test5
| | ${result5} | ${output5}= | Run and Return RC and Output | curl -vv -k -X PUT -d "{\"occupyTimeStamp\": \"09:15\", \"leaveTimeStamp\": \"12:15\", \"duration\": \"180\", \"parkingId\": \"A123\", \"usageServices\": \"Oil\"}" https://localhost:8443/user/vishnu/smartparking 
| | BuiltIn.Should Be Equal As Integers |  ${result5} | 0
| | Should Contain | ${output5} | 200
| Test6
| | ${result6} | ${output6}= | Run and Return RC and Output | curl -vv -k https://localhost:8443/user/vishnu
| | BuiltIn.Should Be Equal As Integers |  ${result6} | 0
| | Should Contain | ${output6} | 200
| Test7
| | ${result7} | ${output7}= | Run and Return RC and Output | curl -vv -k -X PUT -d "{\"Occupied\": \"1\"}" https://128.107.1.74:8443/park/A123
| | BuiltIn.Should Be Equal As Integers |  ${result7} | 0
| | Should Contain | ${output7} | 200
| Test8
| | ${result8} | ${output8}= | Run and Return RC and Output | curl -vv -k -X DELETE https://localhost:8443/user/vishnu 
| | BuiltIn.Should Be Equal As Integers |  ${result8} | 0
| | Should Contain | ${output8} | 200


