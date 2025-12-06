class Solution {
    public boolean isValid(String s) {
        if(s.length()%2!=0) return false;
        Deque<Character>st=new ArrayDeque<>();
        Map<Character, Character>map=Map.of(
                ')','(',
                '}','{',
                ']','['
        );
        for(char ch:s.toCharArray()){
            if(map.containsValue(ch)) st.push(ch);
            else if(map.containsKey(ch) 
            && (st.isEmpty() || st.pop()!=map.get(ch))
            ) 
                return false;
        }

        return st.isEmpty();
    }
}