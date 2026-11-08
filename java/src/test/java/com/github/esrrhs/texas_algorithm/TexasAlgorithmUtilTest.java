package com.github.esrrhs.texas_algorithm;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

public class TexasAlgorithmUtilTest
{
    private static boolean dataAvailable = false;
    private static boolean probDataAvailable = false;

    @BeforeAll
    public static void setUp()
    {
        File colorFile = new File("texas_data_color.txt");
        if (colorFile.exists())
        {
            TexasAlgorithmUtil.load();
            dataAvailable = TexasAlgorithmUtil.isLoaded();
        }

        File winRateFile = new File("texas_data_winrate_0.zst");
        if (winRateFile.exists())
        {
            TexasAlgorithmUtil.loadProbility();
            probDataAvailable = TexasAlgorithmUtil.isProbabilityLoaded();
        }
    }

    @Test
    @DisplayName("Test string and byte conversion utilities")
    public void testConversions()
    {
        String original = "方A,黑K,红10,梅J,黑9";
        List<Byte> pokes = TexasAlgorithmUtil.strToPokes(original);
        assertEquals(5, pokes.size());

        String reconstructed = TexasAlgorithmUtil.pokesToStr(pokes);
        assertEquals("方A黑K红10梅J黑9", reconstructed);

        long key = GenUtil.genCardBind(pokes);
        List<Byte> fromKey = TexasAlgorithmUtil.keyToPoke(key);
        assertEquals(5, fromKey.size());
    }

    @Test
    @DisplayName("Test getMax without wild cards")
    public void testGetMaxWithoutGui()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testGetMaxWithoutGui: data files not present");
            return;
        }
        String best = TexasAlgorithmUtil.getMax("方4,方2", "黑2,黑A,方3,黑5,黑6");
        assertNotNull(best);
        // Best 5 cards should form a straight: 2, 3, 4, 5, 6
        assertTrue(best.contains("2") && best.contains("3") && best.contains("4") && best.contains("5") && best.contains("6"));
    }

    @Test
    @DisplayName("Test table lookup evaluation when data files are available")
    public void testTableLookup()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testTableLookup: data files not present");
            return;
        }

        String cards = "方4,方A,黑2,黑A,黑3,黑5,黑6";
        String cards1 = "红8,方A,方2,黑8,黑3,黑5,黑7";

        int pos = TexasAlgorithmUtil.getWinPosition(cards);
        assertTrue(pos > 0);
        assertEquals(4010, pos);

        double prob = TexasAlgorithmUtil.getWinProbability(cards);
        assertTrue(prob > 0.8 && prob < 1.0);

        int type = TexasAlgorithmUtil.getWinType(cards);
        assertEquals(TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA, type);

        int pos1 = TexasAlgorithmUtil.getWinPosition(cards1);
        assertEquals(1143, pos1);
        int type1 = TexasAlgorithmUtil.getWinType(cards1);
        assertEquals(TexasCardUtil.TEXAS_CARD_TYPE_DUIZI, type1);

        int cmp = TexasAlgorithmUtil.compare(cards, cards1);
        assertTrue(cmp > 0, "cards should beat cards1");
    }

    @Test
    @DisplayName("Test win probability estimation when probability data is available")
    public void testProbabilityEstimation()
    {
        if (!probDataAvailable)
        {
            System.out.println("Skipping testProbabilityEstimation: prob data files not present");
            return;
        }

        float p1 = TexasAlgorithmUtil.getHandProbability("方3,方A", "黑2,黑4,黑5,黑K");
        assertTrue(p1 > 0.55f && p1 < 0.56f, "p1 should be around 0.5533 (exact), got: " + p1);

        float p2 = TexasAlgorithmUtil.getHandProbability("方2,方3", "");
        assertTrue(p2 > 0.35f && p2 < 0.37f, "p2 should be around 0.3598 (exact), got: " + p2);
    }

    @Test
    @DisplayName("Test graceful handling of unloaded probability data")
    public void testUnloadedProbabilityGraceful()
    {
        // A wrong number of hole cards must be reported as an exception, never a crash.
        try
        {
            TexasAlgorithmUtil.getHandProbability(Arrays.asList((byte) 0x02), Collections.emptyList());
            fail("getHandProbability with one hole card should throw");
        }
        catch (IllegalArgumentException expected)
        {
            // expected
        }
    }

    @Test
    @DisplayName("winrate_3 table matches the on-the-fly enumeration")
    public void testWinRate3TableMatchesEnumeration()
    {
        if (!probDataAvailable)
        {
            System.out.println("Skipping testWinRate3TableMatchesEnumeration: prob data files not present");
            return;
        }

        float fromTable = TexasAlgorithmUtil.getHandProbability("方3,梅10", "黑2,红4,方7");

        TexasAlgorithmUtil.WinRateTable saved = TexasAlgorithmUtil.winRateMap[3];
        TexasAlgorithmUtil.winRateMap[3] = null;
        float fromEnum = TexasAlgorithmUtil.getHandProbability("方3,梅10", "黑2,红4,方7");
        TexasAlgorithmUtil.winRateMap[3] = saved;

        assertEquals(fromTable, fromEnum, "winrate_3 table should match the enumeration");
    }

    @Test
    @DisplayName("Missing win-rate tables fall back to enumeration")
    public void testMissingWinRateTableGraceful()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testMissingWinRateTableGraceful: data files not present");
            return;
        }

        TexasAlgorithmUtil.WinRateTable[] saved = new TexasAlgorithmUtil.WinRateTable[6];
        for (int i = 0; i < 6; i++)
        {
            saved[i] = TexasAlgorithmUtil.winRateMap[i];
            TexasAlgorithmUtil.winRateMap[i] = null;
        }

        try
        {
            // Level 3 is used because its enumeration is ~1M evaluations; the
            // preflop one would take minutes.
            float p = TexasAlgorithmUtil.getHandProbability(Arrays.asList((byte) 0x02, (byte) 0x03),
                    Arrays.asList((byte) 0x04, (byte) 0x05, (byte) 0x06));
            assertTrue(p > 0 && p < 1, "flop should be enumerated without tables");

            float q = TexasAlgorithmUtil.getHandProbability(Arrays.asList((byte) 0x02, (byte) 0x03),
                    Arrays.asList((byte) 0x04, (byte) 0x05, (byte) 0x06, (byte) 0x07));
            assertTrue(q > 0 && q < 1, "turn should be enumerated without tables");
        }
        finally
        {
            for (int i = 0; i < 6; i++)
            {
                TexasAlgorithmUtil.winRateMap[i] = saved[i];
            }
        }
    }

    @Test
    @DisplayName("SQLite lookup matches the in-memory table and honors the loaded range")
    public void testSqliteMatchesEnumeration() throws Exception
    {
        File db = new File("texas_data_winrate.db");
        if (!db.exists())
        {
            // The packed database is about 2.3 GiB and is generated locally with
            // go/cmd/packsqlite, so the test skips when it is absent.
            System.out.println("Skipping testSqliteMatchesEnumeration: packed database not present");
            return;
        }

        TexasAlgorithmUtil.loadSqlite(db, 3, 3);
        TexasAlgorithmUtil.WinRateTable saved = TexasAlgorithmUtil.winRateMap[3];
        TexasAlgorithmUtil.winRateMap[3] = null;
        float fromDb;
        try
        {
            fromDb = TexasAlgorithmUtil.getHandProbability("方3,梅10", "黑2,红4,方7");
        }
        finally
        {
            TexasAlgorithmUtil.winRateMap[3] = saved;
        }
        float fromTable = TexasAlgorithmUtil.getHandProbability("方3,梅10", "黑2,红4,方7");
        assertEquals(fromTable, fromDb, 1e-9f, "SQLite row should match the in-memory table");

        // Level 3 is outside 0..2, so the lookup must enumerate; the exact
        // answer still has to equal the table value.
        TexasAlgorithmUtil.loadSqlite(db, 0, 2);
        saved = TexasAlgorithmUtil.winRateMap[3];
        TexasAlgorithmUtil.winRateMap[3] = null;
        float fromEnum;
        try
        {
            fromEnum = TexasAlgorithmUtil.getHandProbability("方3,梅10", "黑2,红4,方7");
        }
        finally
        {
            TexasAlgorithmUtil.winRateMap[3] = saved;
        }
        assertEquals(fromTable, fromEnum, 1e-9f, "a level outside the SQLite range should enumerate to the same value");

        assertThrows(java.io.IOException.class,
                () -> TexasAlgorithmUtil.loadSqlite(new File("texas_data_winrate_missing.db"), 0, 5),
                "a missing database should be rejected");
    }
}
