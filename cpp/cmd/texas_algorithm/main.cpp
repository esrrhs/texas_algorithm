#include <cstdint>
#include <cstring>
#include <fstream>
#include <iostream>
#include <sstream>

#include "gen_extra_opt_util.h"
#include "gen_extra_util.h"
#include "gen_hand_util.h"
#include "gen_opt_util.h"
#include "gen_trans_opt_util.h"
#include "gen_trans_util.h"
#include "texas_algorithm_util.h"

namespace
{

  using namespace texas_algorithm;

  void Compare(const std::string &file);

  // demo mirrors Java's TestUtil.main.
  void Demo()
  {
    Load();

    std::pair<std::string, std::vector<int>> res = GetMaxStr("方2,梅3,黑2,黑4,鬼");
    std::cout << res.first << std::endl;
    std::cout << PokesToStr(res.second) << std::endl;

    res = GetMaxStr("方2,方3,方4,鬼,鬼,黑6,红6");
    std::cout << res.first << std::endl;
    std::cout << PokesToStr(res.second) << std::endl;

    res = GetMaxStr("黑A,方J,黑8,红8,鬼");
    std::cout << res.first << std::endl;
    std::cout << PokesToStr(res.second) << std::endl;

    res = GetMaxStr("方2,鬼,黑2,黑4,黑5,鬼");
    std::cout << res.first << std::endl;
    std::cout << PokesToStr(res.second) << std::endl;

    std::string cards = "方4,方A,黑2,黑A,黑3,黑5,黑6";
    std::string cards1 = "红8,方A,方2,黑8,黑3,黑5,黑7";
    std::cout << GetWinPositionStr(cards) << std::endl;
    std::cout << GetWinProbabilityStr(cards) << std::endl;
    std::cout << KeyToStr(GetWinMaxStr(cards)) << std::endl;
    std::cout << GetWinTypeStr(cards) << std::endl;

    std::cout << GetWinPositionStr(cards1) << std::endl;
    std::cout << GetWinProbabilityStr(cards1) << std::endl;
    std::cout << KeyToStr(GetWinMaxStr(cards1)) << std::endl;
    std::cout << GetWinTypeStr(cards1) << std::endl;

    std::cout << CompareStr(cards, cards1) << std::endl;

    std::string cards2 = "红8,方A,方2,黑8,黑5,黑7";
    std::cout << GetWinPositionStr(cards2) << std::endl;
    std::cout << GetWinProbabilityStr(cards2) << std::endl;
    std::cout << KeyToStr(GetWinMaxStr(cards2)) << std::endl;
    std::cout << GetWinTypeStr(cards2) << std::endl;

    std::string cards3 = "红8,方A,方2,黑5,黑7";
    std::cout << GetWinPositionStr(cards3) << std::endl;
    std::cout << GetWinProbabilityStr(cards3) << std::endl;
    std::cout << KeyToStr(GetWinMaxStr(cards3)) << std::endl;
    std::cout << GetWinTypeStr(cards3) << std::endl;

    std::cout << GetMaxStrHandPub("方4,方2", "黑2,黑A,方3,黑5,黑6").first << std::endl;
    std::cout << GetMaxStrHandPub("方4,方2", "黑2,黑A,黑7,黑5,黑6").first << std::endl;
    std::cout << GetMaxStrHandPub("黑2,黑3", "方2,方A,黑7,黑5,黑6").first << std::endl;
    std::cout << GetMaxStrHandPub("黑2,黑3", "方2,方A,黑7,黑5").first << std::endl;
    std::cout << GetMaxStrHandPub("黑2,黑3", "方2,黑7,黑5").first << std::endl;

    LoadProbility();
    std::cout << GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K") << std::endl;
    std::cout << GetHandProbabilityStr("方2,方3", "") << std::endl;
    std::cout << GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K,方A") << std::endl;

    Compare("hand4/texas_hand_方3方10.txt");
  }

  // compare checks the stored exhaustive win rates against the estimated ones,
  // mirroring Java's TestUtil.compare.
  void Compare(const std::string &file)
  {
    int total = 0;
    int diff1 = 0;
    int diff2 = 0;

    std::ifstream in(file, std::ios::binary);
    if (!in.is_open())
    {
      return;
    }

    std::string line;
    while (std::getline(in, line))
    {
      std::istringstream iss(line);
      std::string keyStr, probStr, card;
      iss >> keyStr >> probStr >> card;
      int64_t key = std::strtoll(keyStr.c_str(), nullptr, 10);
      float probility = std::strtof(probStr.c_str(), nullptr);

      float p = GetHandProbabilityKeyPub(key / 100000000, key % 100000000);

      if (p - probility > 0.1f || probility - p > 0.1f)
      {
        std::cout << "diff " << (p - probility) << " " << card << " " << p << " " << probility << std::endl;
        diff1++;
      }
      if (p - probility > 0.2f || probility - p > 0.2f)
      {
        std::cout << "diff " << (p - probility) << " " << card << " " << p << " " << probility << std::endl;
        diff2++;
      }

      total++;
    }

    std::cout << "diff>0.1 = %" << diff1 * 100 / total << std::endl;
    std::cout << "diff>0.1 = " << diff1 << std::endl;
    std::cout << "diff>0.2 = %" << diff2 * 100 / total << std::endl;
    std::cout << "diff>0.2 = " << diff2 << std::endl;
    std::cout << "total " << total << std::endl;
  }

} // namespace

int main(int argc, char **argv)
{
  if (argc > 1 && std::strcmp(argv[1], "demo") == 0)
  {
    Demo();
    return 0;
  }

  Gen();
  GenExtra();
  GenOpt();
  GenExtraOpt();
  GenTrans();
  GenTransOpt();
  GenHand();
  std::cout << "done!" << std::endl;
  return 0;
}
