/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode(int x) {
 *         val = x;
 *         next = null;
 *     }
 * }
 */
public class Solution {

    public ListNode getIntersectionNode(ListNode headA, ListNode headB){
        ListNode pA=headA, pB=headB;
        while(pA!=pB){
            pA=(pA==null)?headB:pA.next;
            pB=(pB==null)?headA:pB.next;
        }
        return pA;
    }

    // public ListNode getIntersectionNode(ListNode headA, ListNode headB) {
    //     ListNode curA=headA;
    //     ListNode curB=headB;
    //     Set<ListNode>setA=new HashSet<>();
    //     Set<ListNode>setB=new HashSet<>();
    //     while(curA!=null && curB!=null){
    //         if(setA.contains(curB)) return curB;
    //         else if(setB.contains(curA)) return curA;

    //         setA.add(curA);
    //         setB.add(curB);

    //         curA=curA.next;
    //         curB=curB.next;
    //     }

    //     while(curA!=null){
    //         if(setB.contains(curA)) return curA;
        
    //         setB.add(curA);

    //         curA=curA.next;
    //     }

    //     while(curB!=null){
    //         if(setA.contains(curB)) return curB;
            
    //         setA.add(curB);

    //         curB=curB.next;
    //     }

    //     return null;
    // }
}